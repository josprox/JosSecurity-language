package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) evaluateNew(ne *parser.NewExpression) interface{} {
	className := ne.Class.Value

	evalArgs := []interface{}{}
	for _, arg := range ne.Arguments {
		evalArgs = append(evalArgs, r.evaluateExpression(arg))
	}

	// Check PluginRegistry for exported class instantiation first
	if r.PluginRegistry != nil {
		for _, p := range r.PluginRegistry.List() {
			for _, cls := range p.Symbols.Classes {
				if cls.Name == className {
					res, err := r.PluginRegistry.Instantiate(p.Name, cls.Name, evalArgs)
					if err == nil && res != nil {
						return res
					}
					if err != nil {
						panic(fmt.Sprintf("Error instanciando clase de plugin %s::%s: %v", p.Name, cls.Name, err))
					}
				}
			}
		}
	}

	classStmt, ok := r.Classes[className]
	if !ok {
		panic(&JossError{
			Type:    "UndefinedClass",
			Message: fmt.Sprintf("Clase '%s' no encontrada", className),
			File:    r.CurrentFile,
			Line:    ne.Class.Token.Line,
		})
	}

	instance := &Instance{
		Class:  classStmt,
		Fields: make(map[string]interface{}),
	}

	// Collect inheritance chain
	chain := []*parser.ClassStatement{classStmt}
	curr := classStmt
	for curr.SuperClass != nil {
		parentName := curr.SuperClass.Value
		if parent, ok := r.Classes[parentName]; ok {
			chain = append(chain, parent)
			curr = parent
		} else {
			break
		}
	}

	// Initialize properties (Parent -> Child)
	for i := len(chain) - 1; i >= 0; i-- {
		cls := chain[i]
		for _, stmt := range cls.Body.Statements {
			if let, ok := stmt.(*parser.LetStatement); ok {
				instance.Fields[let.Name.Value] = r.evaluateExpression(let.Value)
			}
		}
	}

	// Call constructor if exists
	for _, stmt := range classStmt.Body.Statements {
		if method, ok := stmt.(*parser.MethodStatement); ok {
			if method.Name.Value == "constructor" || method.Name.Value == "main" {
				r.CallMethod(method, instance, ne.Arguments)
				break
			}
		}
		if initStmt, ok := stmt.(*parser.InitStatement); ok {
			if initStmt.Name.Value == "constructor" || initStmt.Name.Value == "main" {
				method := &parser.MethodStatement{
					Token:      initStmt.Token,
					Name:       initStmt.Name,
					Parameters: initStmt.Parameters,
					Body:       initStmt.Body,
				}
				r.CallMethod(method, instance, ne.Arguments)
				break
			}
		}
	}

	return instance
}

func (r *Runtime) evaluateMember(me *parser.MemberExpression) interface{} {
	// 1. Support Static Class Access (e.g. Turnstile::siteKey, GranDB::table, Session::get, etc.)
	if ident, ok := me.Left.(*parser.Identifier); ok {
		className := ident.Value

		// Check user-defined class in r.Classes
		if classStmt, ok := r.Classes[className]; ok {
			if _, native := r.NativeHandlers[className]; native {
				return &BoundMethod{
					Method:      &parser.MethodStatement{Name: &parser.Identifier{Value: me.Property.Value}},
					StaticClass: className,
				}
			}
			for _, stmt := range classStmt.Body.Statements {
				if method, methodOK := stmt.(*parser.MethodStatement); methodOK && method.Name.Value == me.Property.Value {
					return &BoundMethod{Method: method, Instance: &Instance{Class: classStmt, Fields: make(map[string]interface{})}}
				}
			}
		}

		// Check native class
		if r.isNativeClass(className) || IsNativeClass(className) {
			return &BoundMethod{
				Method: &parser.MethodStatement{
					Name: &parser.Identifier{Value: me.Property.Value},
					Body: nil,
				},
				Instance:    nil,
				StaticClass: className,
			}
		}

		// Check loaded plugin in PluginRegistry
		if r.PluginRegistry != nil && r.PluginRegistry.Get(className) != nil {
			return &PluginCallable{
				PluginName: className,
				Function:   me.Property.Value,
			}
		}
	}

	left := r.evaluateExpression(me.Left)

	// Support Plugin Namespace access (e.g. plugin_name::function or plugin_name.function)
	if ns, ok := left.(*PluginNamespace); ok {
		return &PluginCallable{
			PluginName: ns.Name,
			Function:   me.Property.Value,
		}
	}

	// Support Map access via dot notation (e.g. $item.id where $item is a map)
	if m, ok := left.(map[string]interface{}); ok {
		if val, exists := m[me.Property.Value]; exists {
			return val
		}
		return nil
	}

	instance, ok := left.(*Instance)
	if !ok || instance == nil {
		if ident, ok := me.Left.(*parser.Identifier); ok {
			className := ident.Value

			// Check if ident is a loaded plugin name in PluginRegistry
			if r.PluginRegistry != nil && r.PluginRegistry.Get(className) != nil {
				return &PluginCallable{
					PluginName: className,
					Function:   me.Property.Value,
				}
			}

			if classStmt, ok := r.Classes[className]; ok {
				if _, native := r.NativeHandlers[className]; native {
					return &BoundMethod{
						Method:      &parser.MethodStatement{Name: &parser.Identifier{Value: me.Property.Value}},
						StaticClass: className,
					}
				}
				for _, stmt := range classStmt.Body.Statements {
					if method, methodOK := stmt.(*parser.MethodStatement); methodOK && method.Name.Value == me.Property.Value {
						return &BoundMethod{Method: method, Instance: &Instance{Class: classStmt, Fields: make(map[string]interface{})}}
					}
				}
			} else if r.isNativeClass(className) {
				return &BoundMethod{
					Method: &parser.MethodStatement{
						Name: &parser.Identifier{Value: me.Property.Value},
						Body: nil,
					},
					Instance:    nil,
					StaticClass: className,
				}
			}

			// Fallback: check if className is a function/class exported by any loaded plugin
			if r.PluginRegistry != nil {
				for _, p := range r.PluginRegistry.List() {
					for _, fn := range p.Symbols.Functions {
						if fn.Name == me.Property.Value && p.Name == className {
							return &PluginCallable{
								PluginName: p.Name,
								Function:   fn.Name,
							}
						}
					}
					for _, cls := range p.Symbols.Classes {
						if cls.Name == className {
							return &PluginCallable{
								PluginName: p.Name,
								ClassName:  cls.Name,
								Function:   me.Property.Value,
							}
						}
					}
				}
			}

			if left == nil {
				panic(&JossError{
					Type:    "NullReference",
					Message: fmt.Sprintf("Clase o namespace '%s' no encontrado o es nulo", className),
					File:    r.CurrentFile,
					Line:    me.Property.Token.Line,
				})
			}

			panic(&JossError{
				Type:    "UndefinedClass",
				Message: fmt.Sprintf("Clase o plugin '%s' no registrado. Si pertenece a un plugin, verifique joss.yaml y la instalación del paquete", className),
				File:    r.CurrentFile,
				Line:    me.Property.Token.Line,
			})
		}

		if left == nil {
			panic(&JossError{
				Type:    "NullReference",
				Message: fmt.Sprintf("Intento de acceder a la propiedad '%s' sobre un valor nulo o no instanciado", me.Property.Value),
				File:    r.CurrentFile,
				Line:    me.Property.Token.Line,
			})
		}

		panic(&JossError{
			Type:    "NotAnInstance",
			Message: fmt.Sprintf("'%v' (tipo %T) no es una instancia. Intentando acceder a: '%s'", left, left, me.Property.Value),
			File:    r.CurrentFile,
			Line:    me.Property.Token.Line,
		})
	}

	propName := me.Property.Value

	if instance.Fields == nil {
		instance.Fields = make(map[string]interface{})
	}

	if val, ok := instance.Fields[propName]; ok {
		return val
	}

	currentClass := instance.Class
	for currentClass != nil {
		for _, stmt := range currentClass.Body.Statements {
			if method, ok := stmt.(*parser.MethodStatement); ok {
				if method.Name.Value == propName {
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
			if initStmt, ok := stmt.(*parser.InitStatement); ok {
				if initStmt.Name.Value == propName {
					method := &parser.MethodStatement{
						Token:      initStmt.Token,
						Name:       initStmt.Name,
						Parameters: initStmt.Parameters,
						Body:       initStmt.Body,
					}
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
		}

		if currentClass.SuperClass != nil {
			parentName := currentClass.SuperClass.Value
			if parent, ok := r.Classes[parentName]; ok {
				currentClass = parent
			} else {
				currentClass = nil
			}
		} else {
			currentClass = nil
		}
	}

	checkClass := instance.Class
	isNative := false
	for checkClass != nil {
		className := checkClass.Name.Value
		if r.isNativeClass(className) {
			isNative = true
			break
		}
		if checkClass.SuperClass != nil {
			if parent, ok := r.Classes[checkClass.SuperClass.Value]; ok {
				checkClass = parent
			} else {
				break
			}
		} else {
			break
		}
	}

	if isNative {
		return &BoundMethod{
			Method: &parser.MethodStatement{
				Name: &parser.Identifier{Value: propName},
				Body: nil,
			},
			Instance: instance,
		}
	}

	if r.PluginRegistry != nil && instance.Class != nil {
		className := instance.Class.Name.Value
		for _, p := range r.PluginRegistry.List() {
			for _, cls := range p.Symbols.Classes {
				if cls.Name == className {
					for _, m := range cls.Methods {
						if m.Name == propName {
							return &BoundMethod{
								Method: &parser.MethodStatement{
									Name: &parser.Identifier{Value: propName},
									Body: nil,
								},
								Instance: instance,
							}
						}
					}
				}
			}
		}
	}

	className := "Anonymous"
	if instance.Class != nil {
		className = instance.Class.Name.Value
	}
	panic(&JossError{
		Type:    "UndefinedProperty",
		Message: fmt.Sprintf("Propiedad o método '%s' no encontrado en clase '%s'", propName, className),
		File:    r.CurrentFile,
		Line:    me.Property.Token.Line,
	})
}
