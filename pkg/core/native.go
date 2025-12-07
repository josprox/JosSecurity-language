package core

import (
	"github.com/jossecurity/joss/pkg/parser"
)

// Helper to register a native class and its handler
func (r *Runtime) registerNative(name string, methods []string, handler NativeHandler) {
	// Build MethodStatements
	stmts := []parser.Statement{}
	for _, m := range methods {
		stmts = append(stmts, &parser.MethodStatement{Name: &parser.Identifier{Value: m}})
	}

	classStmt := &parser.ClassStatement{
		Name: &parser.Identifier{Value: name},
		Body: &parser.BlockStatement{Statements: stmts},
	}
	r.registerClass(classStmt)
	r.NativeHandlers[name] = handler
}

// RegisterNativeClasses injects the native class definitions into the runtime
func (r *Runtime) RegisterNativeClasses() {
	// Note: We use (*Runtime).methodName to register UNBOUND methods as handlers.
	// This ensures that when they are called, we pass the *current* execution runtime 'r',
	// not the original 'r' that tried to register them.

	// Stack
	r.registerNative("Stack", []string{}, (*Runtime).executeStackMethod)

	// Queue
	r.registerNative("Queue", []string{}, (*Runtime).executeQueueMethod)

	// GranDB
	r.registerNative("GranDB", []string{}, (*Runtime).executeGranMySQLMethod)
	// Alias for compatibility
	r.registerNative("GranMySQL", []string{"table", "select", "where", "innerJoin", "leftJoin", "rightJoin", "get", "first", "insert", "update", "delete", "deleteAll", "truncate", "query", "orderBy", "limit", "offset", "count"}, (*Runtime).executeGranMySQLMethod)
	r.NativeHandlers["GranMySQL"] = (*Runtime).executeGranMySQLMethod

	// Auth
	r.registerNative("Auth", []string{"user", "check", "guest", "id", "logout", "attempt", "create", "hasRole", "verify", "refresh", "delete"}, (*Runtime).executeAuthMethod)
	r.Variables["Auth"] = &Instance{Class: r.Classes["Auth"], Fields: make(map[string]interface{})}

	// System
	r.registerNative("System", []string{"env", "Run", "load_driver"}, (*Runtime).executeSystemMethod)
	r.Variables["System"] = &Instance{Class: r.Classes["System"], Fields: make(map[string]interface{})}

	// SmtpClient
	r.registerNative("SmtpClient", []string{"auth", "secure", "send"}, (*Runtime).executeSmtpClientMethod)

	// Cron
	r.registerNative("Cron", []string{"schedule"}, (*Runtime).executeCronMethod)
	r.Variables["Cron"] = &Instance{Class: r.Classes["Cron"], Fields: make(map[string]interface{})}

	// Task
	r.registerNative("Task", []string{"on_request"}, (*Runtime).executeTaskMethod)
	r.Variables["Task"] = &Instance{Class: r.Classes["Task"], Fields: make(map[string]interface{})}

	// View
	r.registerNative("View", []string{"render"}, (*Runtime).executeViewMethod)
	r.Variables["View"] = &Instance{Class: r.Classes["View"], Fields: make(map[string]interface{})}

	// Router
	r.registerNative("Router", []string{"get", "post", "put", "delete", "match", "api", "group", "middleware", "end"}, (*Runtime).executeRouterMethod)
	r.Variables["Router"] = &Instance{Class: r.Classes["Router"], Fields: make(map[string]interface{})}

	// Request
	r.registerNative("Request", []string{"input", "post", "all", "except", "get", "file"}, (*Runtime).executeRequestMethod)
	r.Variables["Request"] = &Instance{Class: r.Classes["Request"], Fields: make(map[string]interface{})}

	// Response
	r.registerNative("Response", []string{"json", "redirect", "error", "raw"}, (*Runtime).executeResponseMethod)
	r.Variables["Response"] = &Instance{Class: r.Classes["Response"], Fields: make(map[string]interface{})}

	// RedirectResponse
	r.registerNative("RedirectResponse", []string{}, (*Runtime).executeRedirectResponseMethod)

	// WebSocket
	r.registerNative("WebSocket", []string{"broadcast"}, (*Runtime).executeWebSocketMethod)
	r.Variables["WebSocket"] = &Instance{Class: r.Classes["WebSocket"], Fields: make(map[string]interface{})}

	// Schema
	r.registerNative("Schema", []string{"create", "table"}, (*Runtime).executeSchemaMethod)
	r.Variables["Schema"] = &Instance{Class: r.Classes["Schema"], Fields: make(map[string]interface{})}

	// Blueprint
	r.registerNative("Blueprint", []string{}, (*Runtime).executeBlueprintMethod)

	// Redis
	r.registerNative("Redis", []string{}, (*Runtime).executeRedisMethod)
	r.Variables["Redis"] = &Instance{Class: r.Classes["Redis"], Fields: make(map[string]interface{})}

	// Migration
	r.registerNative("Migration", []string{}, nil) // No direct native method handler probably, or handled via Schema?

	// Math
	r.registerNative("Math", []string{"random", "floor", "ceil", "abs"}, (*Runtime).executeMathMethod)
	r.Variables["Math"] = &Instance{Class: r.Classes["Math"], Fields: make(map[string]interface{})}

	// Session
	r.registerNative("Session", []string{"get", "put", "has", "forget", "all"}, (*Runtime).executeSessionMethod)
	// Session is instantiated per request

	// UUID
	r.registerNative("UUID", []string{"generate", "v4"}, (*Runtime).executeUUIDMethod)
	r.Variables["UUID"] = &Instance{Class: r.Classes["UUID"], Fields: make(map[string]interface{})}

	// UserStorage
	r.registerNative("UserStorage", []string{"put", "get", "update", "path", "exists", "delete"}, (*Runtime).executeUserStorageMethod)
	r.Variables["UserStorage"] = &Instance{Class: r.Classes["UserStorage"], Fields: make(map[string]interface{})}
}

func (r *Runtime) executeNativeMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Traverse class hierarchy (bottom-up)
	currentClass := instance.Class
	for currentClass != nil {
		className := currentClass.Name.Value

		// Optimize: Check simple map lookup
		if handler, ok := r.NativeHandlers[className]; ok {
			if handler != nil {
				// PASS 'r' (the current runtime) as the first argument
				return handler(r, instance, method, args)
			}
		}

		// Move to parent
		if currentClass.SuperClass != nil {
			if parent, ok := r.Classes[currentClass.SuperClass.Value]; ok {
				currentClass = parent
			} else {
				break
			}
		} else {
			break
		}
	}
	return nil
}
