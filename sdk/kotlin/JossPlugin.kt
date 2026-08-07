package joss.sdk

import java.nio.charset.StandardCharsets

/**
 * Idiomatic Kotlin DSL Framework for Joss JP v2 sidecars (joss-rpc-v1).
 * Supports Kotlin/JVM and Kotlin/Native for single-binary plugins.
 */
class JossPlugin private constructor() {
    companion object {
        const val PROTOCOL = "joss-rpc-v1"

        fun define(block: JossPluginBuilder.() -> Unit): JossPluginRunner {
            val builder = JossPluginBuilder()
            builder.block()
            return builder.build()
        }

        fun run(dispatch: (String) -> String) {
            try {
                val rawInput = System.`in`.readBytes().toString(StandardCharsets.UTF_8)
                val response = dispatch(rawInput)
                println(response)
            } catch (e: Throwable) {
                val errorJson = """{"id":"","error":{"code":"${e.javaClass.simpleName}","message":"${escapeJson(e.message)}"}}"""
                println(errorJson)
            }
        }

        private fun escapeJson(raw: String?): String {
            if (raw == null) return ""
            return raw.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n").replace("\r", "\\r")
        }
    }
}

class JossPluginBuilder {
    private val handlers = mutableMapOf<String, (String) -> String>()

    fun method(name: String, handler: (String) -> String) {
        handlers[name] = handler
    }

    fun build(): JossPluginRunner {
        return JossPluginRunner(handlers)
    }
}

class JossPluginRunner(private val handlers: Map<String, (String) -> String>) {
    fun run() {
        JossPlugin.run { rawRequest ->
            // Execute method or return error
            handlers.values.firstOrNull()?.invoke(rawRequest) ?: """{"id":"","error":{"code":"UNKNOWN_METHOD","message":"No handler registered"}}"""
        }
    }
}
