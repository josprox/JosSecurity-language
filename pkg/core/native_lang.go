package core

import (
	"github.com/jossecurity/joss/pkg/i18n"
)

// Lang Native Class
// Lang::get("welcome", {"name": "Juan"})
// Lang::set("es")
// Lang::locale() -> "es"

func (r *Runtime) executeLangMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "set":
		if len(args) > 0 {
			if loc, ok := args[0].(string); ok {
				r.SetLocale(loc)
				return true
			}
		}
	case "get":
		// Args: key (string), replacements (map, optional)
		if len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return ""
			}

			replacements := make(map[string]interface{})
			if len(args) > 1 {
				if rMap, ok := args[1].(*Instance); ok && rMap.Class.Name.Value == "Map" {
					// Convert native Map instance to go map?
					// Wait, literal maps in Joss are passed as...
					// In JosSecurity, map literals `{}` are mostly usually `*Instance` of Map class or strictly `map[string]interface{}`?
					// Let's check how Map literals are handled.
					// In `evalMapLiteral`, it returns `&Instance{Class: r.Classes["Map"], Fields: ...}`? Or `map[string]interface{}`?
					// Usually primitive maps are not instances.
					// Let's assume generic interface{} map or Instance.
					// If it's a native Map instance, we iterate its Fields.
					for k, v := range rMap.Fields {
						replacements[k] = v
					}
				} else if rMap, ok := args[1].(map[string]interface{}); ok {
					replacements = rMap
				}
			}

			// Use runtime's current locale
			locale := r.GetLocale()
			return i18n.GlobalManager.Get(locale, key, replacements)
		}
	case "locale":
		return r.GetLocale()

	case "locales":
		return i18n.GlobalManager.GetAvailableLocales()
	}
	return nil
}

// Helper on Runtime to manage locale state per request/runtime
func (r *Runtime) SetLocale(l string) {
	// Store locale in a special variable or field
	// Since Runtime struct is in runtime.go, we can't add field here easily without editing struct.
	// We can use r.Variables["_LOCALE"] or similar.
	r.Variables["_LOCALE"] = l
}

func (r *Runtime) GetLocale() string {
	if val, ok := r.Variables["_LOCALE"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return i18n.GlobalManager.DefaultLocale // fallback
}
