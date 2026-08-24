package core

import "sync"

var (
	builtinNamesOnce sync.Once
	builtinNamesMap  map[string]bool
)

// builtinList holds all registered Joss built-in functions across date, string, array, async, and IO modules.
var builtinList = []string{
	// Output & Execution
	"echo", "print", "exit", "die",
	// Language constructs
	"isset", "empty", "len",
	// Date & Time
	"date", "time", "sleep", "microtime",
	// String functions
	"str_upper", "str_lower", "str_contains", "str_replace", "str_trim",
	"str_split", "str_join", "substr", "strlen", "strpos", "sprintf",
	// Array & Map functions
	"array_push", "array_pop", "array_shift", "array_unshift",
	"array_keys", "array_values", "array_merge", "in_array", "count",
	// Async & Channels
	"async", "await", "make_chan", "send", "recv", "close",
	// IO, HTTP & Utilities
	"json_encode", "json_decode", "file_get_contents", "file_put_contents",
	"env", "config", "json", "redirect", "back", "view", "response", "request",
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
