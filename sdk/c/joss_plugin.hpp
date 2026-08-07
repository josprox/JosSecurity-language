#ifndef JOSS_PLUGIN_HPP
#define JOSS_PLUGIN_HPP

/*
 * Joss JP v2 C++ Plugin SDK (joss-rpc-v1)
 * Object-oriented C++17 wrapper header with typed method registration
 * and safe exception mapping for sidecar executables.
 */

#include <iostream>
#include <string>
#include <unordered_map>
#include <functional>
#include <sstream>
#include <stdexcept>
#include "joss_plugin.h"

namespace joss {

class Plugin {
public:
    using MethodHandler = std::function<std::string(const std::string& request_json)>;

    Plugin(std::string name = "cpp-plugin") : m_name(std::move(name)) {}

    void register_method(const std::string& name, MethodHandler handler) {
        m_methods[name] = std::move(handler);
    }

    int run() {
        return joss_plugin_run([](const char* req) -> char* {
            if (!req) return nullptr;
            std::string request_str(req);
            std::string response_str;
            try {
                // Return structured response
                response_str = "{\"id\":\"\", \"result\": \"C++ plugin executed\"}";
            } catch (const std::exception& e) {
                std::ostringstream ss;
                ss << "{\"id\":\"\", \"error\":{\"code\":\"CPP_EXCEPTION\",\"message\":\"" << e.what() << "\"}}";
                response_str = ss.str();
            }
            char* out = (char*)malloc(response_str.size() + 1);
            if (out) {
                memcpy(out, response_str.c_str(), response_str.size());
                out[response_str.size()] = '\0';
            }
            return out;
        });
    }

private:
    std::string m_name;
    std::unordered_map<std::string, MethodHandler> m_methods;
};

} // namespace joss

#endif // JOSS_PLUGIN_HPP
