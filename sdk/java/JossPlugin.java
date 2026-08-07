package joss.sdk;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Function;

/**
 * Robust Java SDK for Joss JP v2 sidecars (joss-rpc-v1).
 * Compatible with standard JVM and GraalVM native-image.
 */
public final class JossPlugin {
    public static final String PROTOCOL = "joss-rpc-v1";

    private final Map<String, Function<String, String>> methods = new HashMap<>();

    public JossPlugin registerMethod(String name, Function<String, String> handler) {
        methods.put(name, handler);
        return this;
    }

    public static void run(Function<String, String> dispatch) throws IOException {
        String request = new String(System.in.readAllBytes(), StandardCharsets.UTF_8);
        String response;
        try {
            response = dispatch.apply(request);
            if (response == null || response.isBlank()) {
                response = "{\"id\":\"\",\"error\":{\"code\":\"EMPTY_RESPONSE\",\"message\":\"Handler returned empty response\"}}";
            }
        } catch (Exception e) {
            response = String.format(
                "{\"id\":\"\",\"error\":{\"code\":\"%s\",\"message\":\"%s\"}}",
                e.getClass().getSimpleName(),
                escapeJson(e.getMessage() != null ? e.getMessage() : e.toString())
            );
        }
        System.out.write(response.getBytes(StandardCharsets.UTF_8));
        System.out.write('\n');
        System.out.flush();
    }

    private static String escapeJson(String raw) {
        if (raw == null) return "";
        return raw.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n").replace("\r", "\\r");
    }
}
