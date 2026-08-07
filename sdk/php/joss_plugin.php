<?php

declare(strict_types=1);

const JOSS_PLUGIN_PROTOCOL = 'joss-rpc-v1';

/**
 * Robust OOP Framework for Joss JP v2 native sidecars (joss-rpc-v1).
 */
class JossPlugin
{
    private string $name;
    private string $version;
    /** @var array<string, callable> */
    private array $methods = [];

    public function __construct(string $name = 'php-plugin', string $version = '1.0.0')
    {
        $this->name = $name;
        $this->version = $version;
    }

    public function register(string $methodName, callable $handler): self
    {
        $this->methods[$methodName] = $handler;
        return $this;
    }

    public function run(): void
    {
        $requestId = '';

        try {
            $raw = stream_get_contents(STDIN);
            if ($raw === false || trim($raw) === '') {
                throw new InvalidArgumentException('Empty request received from STDIN');
            }

            $request = json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
            if (!is_array($request)) {
                throw new InvalidArgumentException('Request must be a JSON object');
            }

            $requestId = isset($request['id']) ? (string) $request['id'] : '';
            if (($request['protocol'] ?? null) !== JOSS_PLUGIN_PROTOCOL) {
                throw new InvalidArgumentException('Unsupported protocol version: ' . ($request['protocol'] ?? 'null'));
            }

            $method = isset($request['method']) ? (string) $request['method'] : '';
            if ($method === '' || !isset($this->methods[$method]) || !is_callable($this->methods[$method])) {
                throw new BadMethodCallException('Unknown method: ' . $method);
            }

            $args = $request['args'] ?? [];
            if (!is_array($args) || !array_is_list($args)) {
                throw new InvalidArgumentException('args must be a JSON array');
            }

            $result = ($this->methods[$method])(...$args);

            if ($result instanceof Generator) {
                foreach ($result as $chunk) {
                    $this->writeFrame(['id' => $requestId, 'event' => 'chunk', 'content' => $chunk]);
                }
                $result = $result->getReturn();
            }

            $this->writeFrame(['id' => $requestId, 'result' => $result]);
        } catch (Throwable $error) {
            $this->writeFrame([
                'id' => $requestId,
                'error' => [
                    'code' => (new ReflectionClass($error))->getShortName(),
                    'message' => $error->getMessage(),
                ],
            ]);
        }
    }

    /** @param array<string, mixed> $frame */
    private function writeFrame(array $frame): void
    {
        fwrite(STDOUT, json_encode($frame, JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES) . PHP_EOL);
        fflush(STDOUT);
    }
}

/** Legacy global function helper */
function runJossPlugin(array $methods): void
{
    $plugin = new JossPlugin();
    foreach ($methods as $name => $handler) {
        $plugin->register($name, $handler);
    }
    $plugin->run();
}
