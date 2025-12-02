package core

import (
	"github.com/jossecurity/joss/pkg/parser"
)

// RegisterNativeClasses injects the native class definitions into the runtime
func (r *Runtime) RegisterNativeClasses() {
	// Stack
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "Stack"},
		Body: &parser.BlockStatement{},
	})

	// Queue
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "Queue"},
		Body: &parser.BlockStatement{},
	})

	// GranDB (Database Abstraction)
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "GranDB"},
		Body: &parser.BlockStatement{},
	})

	// Auth
	authClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Auth"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "user"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "check"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "guest"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "id"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "logout"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "attempt"}},
			},
		},
	}
	r.registerClass(authClass)
	r.Variables["Auth"] = &Instance{Class: authClass, Fields: make(map[string]interface{})}

	// System
	systemClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "System"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(systemClass)
	r.Variables["System"] = &Instance{Class: systemClass, Fields: make(map[string]interface{})}

	// SmtpClient
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "SmtpClient"},
		Body: &parser.BlockStatement{},
	})

	// Cron
	cronClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Cron"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(cronClass)
	r.Variables["Cron"] = &Instance{Class: cronClass, Fields: make(map[string]interface{})}

	// Task
	taskClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Task"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(taskClass)
	r.Variables["Task"] = &Instance{Class: taskClass, Fields: make(map[string]interface{})}

	// View
	viewClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "View"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "render"}},
			},
		},
	}
	r.registerClass(viewClass)
	r.Variables["View"] = &Instance{Class: viewClass, Fields: make(map[string]interface{})}

	// Router
	routerClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Router"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "get"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "post"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "put"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "delete"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "match"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "api"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "group"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "middleware"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "end"}},
			},
		},
	}
	r.registerClass(routerClass)
	r.Variables["Router"] = &Instance{Class: routerClass, Fields: make(map[string]interface{})}

	// Request
	requestClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Request"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "input"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "post"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "all"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "except"}},
			},
		},
	}
	r.registerClass(requestClass)
	r.Variables["Request"] = &Instance{Class: requestClass, Fields: make(map[string]interface{})}

	// Response
	responseClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Response"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(responseClass)
	r.Variables["Response"] = &Instance{Class: responseClass, Fields: make(map[string]interface{})}

	// RedirectResponse (Helper for chaining)
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "RedirectResponse"},
		Body: &parser.BlockStatement{},
	})

	// WebSocket
	wsClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "WebSocket"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(wsClass)
	r.Variables["WebSocket"] = &Instance{Class: wsClass, Fields: make(map[string]interface{})}

	// Schema
	schemaClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Schema"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(schemaClass)
	r.Variables["Schema"] = &Instance{Class: schemaClass, Fields: make(map[string]interface{})}

	// Blueprint (for Schema migrations)
	blueprintClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Blueprint"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(blueprintClass)

	// Redis
	redisClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Redis"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(redisClass)
	r.Variables["Redis"] = &Instance{Class: redisClass, Fields: make(map[string]interface{})}
	// Migration
	migrationClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Migration"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(migrationClass)
	r.registerClass(migrationClass)

	// Math
	mathClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Math"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "random"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "floor"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "ceil"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "abs"}},
			},
		},
	}
	r.registerClass(mathClass)
	r.Variables["Math"] = &Instance{Class: mathClass, Fields: make(map[string]interface{})}

	// Session
	sessionClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Session"},
		Body: &parser.BlockStatement{
			Statements: []parser.Statement{
				&parser.MethodStatement{Name: &parser.Identifier{Value: "get"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "put"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "has"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "forget"}},
				&parser.MethodStatement{Name: &parser.Identifier{Value: "all"}},
			},
		},
	}
	r.registerClass(sessionClass)
	// Session is instantiated per request, but we register the class here.
}

func (r *Runtime) executeNativeMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Check class and super classes for native handler
	currentClass := instance.Class
	for currentClass != nil {
		switch currentClass.Name.Value {
		case "Stack":
			return r.executeStackMethod(instance, method, args)
		case "Queue":
			return r.executeQueueMethod(instance, method, args)
		case "GranMySQL", "GranDB":
			return r.executeGranMySQLMethod(instance, method, args)
		case "Auth":
			return r.executeAuthMethod(instance, method, args)
		case "System":
			return r.executeSystemMethod(instance, method, args)
		case "SmtpClient":
			return r.executeSmtpClientMethod(instance, method, args)
		case "Cron":
			return r.executeCronMethod(instance, method, args)
		case "Task":
			return r.executeTaskMethod(instance, method, args)
		case "View":
			return r.executeViewMethod(instance, method, args)
		case "Router":
			return r.executeRouterMethod(instance, method, args)
		case "Request":
			return r.executeRequestMethod(instance, method, args)
		case "Response":
			return r.executeResponseMethod(instance, method, args)
		case "RedirectResponse":
			return r.executeRedirectResponseMethod(instance, method, args)
		case "WebSocket":
			return r.executeWebSocketMethod(instance, method, args)
		case "Schema":
			return r.executeSchemaMethod(instance, method, args)
		case "Blueprint":
			return r.executeBlueprintMethod(instance, method, args)
		case "Redis":
			return r.executeRedisMethod(instance, method, args)
		case "Math":
			return r.executeMathMethod(instance, method, args)
		case "Session":
			return r.executeSessionMethod(instance, method, args)
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
