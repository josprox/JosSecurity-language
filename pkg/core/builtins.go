package core

import "sync"

var (
	builtinNamesOnce sync.Once
	builtinNamesMap  map[string]bool
)

// builtinList holds all registered Joss built-in functions across date, string, array, async, IO, and web framework helpers.
var builtinList = []string{
	// Output & Execution
	"echo", "print", "exit", "die", "dump", "dd", "panic", "var_dump", "print_r",
	// Language constructs & checks
	"isset", "empty", "len", "is_null", "is_string", "is_numeric", "is_int", "is_bool", "is_object", "is_array", "file_exists",
	"get_class", "method_exists", "property_exists",
	// Date & Time
	"date", "time", "now", "sleep", "microtime", "strtotime",
	// String functions
	"str_upper", "str_lower", "str_contains", "str_replace", "str_trim",
	"str_split", "str_join", "substr", "strlen", "strpos", "sprintf",
	"strtolower", "strtoupper", "str_ends_with", "str_starts_with", "trim", "md5", "sha1",
	// Array & Map functions
	"array_push", "array_pop", "array_shift", "array_unshift",
	"array_keys", "array_values", "array_merge", "in_array", "count", "intval", "end", "implode", "explode",
	"head", "tail", "range",
	// Async & Channels
	"async", "await", "make_chan", "send", "recv", "close",
	// IO, HTTP, Web & Framework Utilities
	"json_encode", "json_decode", "file_get_contents", "file_put_contents",
	"env", "config", "json", "redirect", "back", "view", "response", "request",
	"csrf_field", "csrf_token", "route", "url", "asset", "session", "auth",
}

func initBuiltinMap() {
	builtinNamesOnce.Do(func() {
		builtinNamesMap = make(map[string]bool, len(builtinList))
		for _, name := range builtinList {
			builtinNamesMap[name] = true
		}
	})
}

// IsBuiltin returns true if the function name is a core built-in function in Joss.
func IsBuiltin(name string) bool {
	initBuiltinMap()
	return builtinNamesMap[name]
}

// GetBuiltinFunctionNames returns a list of all built-in function names.
func GetBuiltinFunctionNames() []string {
	initBuiltinMap()
	result := make([]string, len(builtinList))
	copy(result, builtinList)
	return result
}
