const vscode = require('vscode');

function activate(context) {
    console.log('JosSecurity Extension is now active!');

    // 1. Autocomplete Provider
    const provider = vscode.languages.registerCompletionItemProvider('joss', {
        provideCompletionItems(document, position, token, context) {
            const linePrefix = document.lineAt(position).text.substr(0, position.character);
            
            // Class-based Autocomplete (Static methods with ::)
            if (linePrefix.endsWith('Security::')) {
                return [
                    createMethod('loadEnv', 'loadEnv()', 'Loads and decrypts environment variables.')
                ];
            }
            if (linePrefix.endsWith('Server::')) {
                return [
                    createMethod('config', 'config(options)', 'Configures the web server.'),
                    createMethod('loadRoutes', 'loadRoutes(file)', 'Loads HTML routes from file.'),
                    createMethod('loadApi', 'loadApi(file)', 'Loads API routes from file.'),
                    createMethod('start', 'start()', 'Starts the web server.')
                ];
            }
            if (linePrefix.endsWith('Log::')) {
                return [
                    createMethod('info', 'info(msg)', 'Logs an informational message.'),
                    createMethod('error', 'error(msg)', 'Logs an error message.')
                ];
            }
            if (linePrefix.endsWith('Router::')) {
                return [
                    createMethod('get', 'get(path, handler)', 'Defines a GET route.'),
                    createMethod('post', 'post(path, handler)', 'Defines a POST route.'),
                    createMethod('put', 'put(path, handler)', 'Defines a PUT route.'),
                    createMethod('delete', 'delete(path, handler)', 'Defines a DELETE route.')
                ];
            }
            if (linePrefix.endsWith('Task::')) {
                return [
                    createMethod('on_request', 'on_request(name, interval, callback)', 'Schedules a hit-based task.')
                ];
            }
            if (linePrefix.endsWith('Auth::')) {
                return [
                    createMethod('create', 'create(data)', 'Creates a new user with hashed password.'),
                    createMethod('attempt', 'attempt(email, password)', 'Attempts login and returns JWT token.'),
                    createMethod('refresh', 'refresh(token)', 'Refreshes a JWT token.')
                ];
            }
            
            // Instance methods (with .)
            if (linePrefix.endsWith('System.')) {
                return [
                    createMethod('Run', 'Run(command, args)', 'Executes a system command.'),
                    createMethod('env', 'env(key, default)', 'Retrieves an environment variable securely.')
                ];
            }
            if (linePrefix.endsWith('SmtpClient.')) {
                return [
                    createMethod('auth', 'auth(user, pass)', 'Sets SMTP credentials.'),
                    createMethod('secure', 'secure(mode)', 'Sets security mode (TLS/SSL).'),
                    createMethod('send', 'send(to, subject, body)', 'Sends an email.')
                ];
            }
            if (linePrefix.endsWith('Cron.')) {
                return [
                    createMethod('schedule', 'schedule(name, time, callback)', 'Schedules a daemon task.')
                ];
            }
            if (linePrefix.endsWith('View.')) {
                return [
                    createMethod('render', 'render(viewName, data)', 'Renders an HTML view.')
                ];
            }

            // Global Functions
            return [
                createFunction('toon_encode', 'toon_encode(data)', 'Encodes data to TOON format.'),
                createFunction('toon_decode', 'toon_decode(string)', 'Decodes TOON string to data.'),
                createFunction('toon_verify', 'toon_verify(string)', 'Verifies TOON string structure.'),
                createFunction('json_encode', 'json_encode(data)', 'Encodes data to JSON.'),
                createFunction('json_decode', 'json_decode(string)', 'Decodes JSON string to data.'),
                createFunction('json_verify', 'json_verify(string)', 'Verifies JSON string validity.'),
                createFunction('print', 'print(msg)', 'Prints to stdout.'),
                createFunction('echo', 'echo(msg)', 'Alias for print.'),
                createFunction('env', 'env(key, default)', 'Global alias for System.env.'),
                createKeyword('class', 'Defines a new class.'),
                createKeyword('function', 'Defines a new function.'),
                createKeyword('Init', 'Defines an initializer/constructor.'),
                createKeyword('async', 'Starts an asynchronous operation.'),
                createKeyword('await', 'Waits for an asynchronous operation.'),
            ];
        }
    }, '.', ':'); // Trigger on '.' and ':'

    // 2. Hover Provider
    const hoverProvider = vscode.languages.registerHoverProvider('joss', {
        provideHover(document, position, token) {
            const range = document.getWordRangeAtPosition(position);
            const word = document.getText(range);

            if (word === 'System') return new vscode.Hover('**System Class**\n\nProvides access to system commands and environment variables.');
            if (word === 'SmtpClient') return new vscode.Hover('**SmtpClient Class**\n\nNative SMTP client for sending emails.');
            if (word === 'Cron') return new vscode.Hover('**Cron Class**\n\nScheduler for daemon tasks.');
            if (word === 'Task') return new vscode.Hover('**Task Class**\n\nScheduler for hit-based tasks.');
            if (word === 'View') return new vscode.Hover('**View Class**\n\nRender engine for HTML views.');
            if (word === 'toon_encode') return new vscode.Hover('**toon_encode(data)**\n\nConverts an array or map to TOON format.');
            
            return null;
        }
    });

    context.subscriptions.push(provider, hoverProvider);
}

function createMethod(label, detail, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Method);
    item.detail = detail;
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

function createFunction(label, detail, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Function);
    item.detail = detail;
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

function createKeyword(label, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Keyword);
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

exports.activate = activate;

function deactivate() {}

exports.deactivate = deactivate;
