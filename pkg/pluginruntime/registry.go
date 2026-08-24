package pluginruntime

import (
	"fmt"
	"sync"
)

// PluginRegistry almacena los plugins cargados y resuelve invocaciones.
type PluginRegistry struct {
	plugins map[string]*Plugin
	mu      sync.RWMutex
	host    HostContext
}

// NewPluginRegistry inicializa un nuevo registro de plugins.
func NewPluginRegistry(host HostContext) *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*Plugin),
		host:    host,
	}
}

// Register registra un plugin validado en el sistema.
func (r *PluginRegistry) Register(plugin *Plugin) error {
	if plugin == nil {
		return fmt.Errorf("pluginruntime: intento de registrar plugin nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[plugin.Name]; exists {
		return fmt.Errorf("%w: %s v%s", ErrPluginAlreadyLoaded, plugin.Name, plugin.Version)
	}

	r.plugins[plugin.Name] = plugin
	return nil
}

// Get obtiene un plugin por su nombre.
func (r *PluginRegistry) Get(pluginName string) *Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.plugins[pluginName]
}

// List retorna todos los plugins registrados.
func (r *PluginRegistry) List() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}

// CallFunction ejecuta una funcion exportada por un plugin de forma segura y aislada.
func (r *PluginRegistry) CallFunction(pluginName, fnName string, args []interface{}) (res interface{}, err error) {
	plugin := r.Get(pluginName)
	if plugin == nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotFound, pluginName)
	}

	return SafeCall(pluginName, fnName, func() (interface{}, error) {
		switch plugin.Format {
		case FormatJossAST:
			if plugin.jossExecutor == nil {
				return nil, fmt.Errorf("plugin %s no tiene ejecutor AST disponible", pluginName)
			}
			return plugin.jossExecutor.CallFunction(fnName, args)

		case FormatJPBC:
			guard := NewPermissionGuard(plugin.Metadata.Permissions)
			vm := NewJPBCVM(plugin.jpbcModule, guard, r.host)
			return vm.Execute(fnName, args)

		default:
			return nil, fmt.Errorf("formato no soportado para ejecucion: %s", plugin.Format)
		}
	})
}

// Instantiate crea una instancia de una clase exportada por un plugin.
func (r *PluginRegistry) Instantiate(pluginName, className string, args []interface{}) (res interface{}, err error) {
	plugin := r.Get(pluginName)
	if plugin == nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotFound, pluginName)
	}

	return SafeCall(pluginName, className+".init", func() (interface{}, error) {
		switch plugin.Format {
		case FormatJossAST:
			if plugin.jossExecutor == nil {
				return nil, fmt.Errorf("plugin %s no tiene ejecutor AST disponible", pluginName)
			}
			return plugin.jossExecutor.Instantiate(className, args)

		case FormatJPBC:
			// En JPBC, las estructuras se instancian como mapas tipados
			instanceMap := make(map[string]interface{})
			instanceMap["__type__"] = className
			instanceMap["__plugin__"] = pluginName
			
			// Look for and execute constructor if it exists
			initName := fmt.Sprintf("%s/init", className)
			if _, exists := plugin.jpbcModule.Functions[initName]; exists {
				guard := NewPermissionGuard(plugin.Metadata.Permissions)
				vm := NewJPBCVM(plugin.jpbcModule, guard, r.host)
				initArgs := append([]interface{}{instanceMap}, args...)
				_, err := vm.Execute(initName, initArgs)
				if err != nil {
					return nil, fmt.Errorf("error al inicializar instancia: %w", err)
				}
			}

			return instanceMap, nil

		default:
			return nil, fmt.Errorf("formato no soportado para instanciacion: %s", plugin.Format)
		}
	})
}

// CallMethod ejecuta un metodo sobre una instancia creada previamente por un plugin.
func (r *PluginRegistry) CallMethod(pluginName, className, methodName string, instance interface{}, args []interface{}) (res interface{}, err error) {
	plugin := r.Get(pluginName)
	if plugin == nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotFound, pluginName)
	}

	return SafeCall(pluginName, className+"."+methodName, func() (interface{}, error) {
		switch plugin.Format {
		case FormatJossAST:
			if plugin.jossExecutor == nil {
				return nil, fmt.Errorf("plugin %s no tiene ejecutor AST disponible", pluginName)
			}
			return plugin.jossExecutor.CallMethod(instance, methodName, args)

		case FormatJPBC:
			qualifiedName := fmt.Sprintf("%s/%s", className, methodName)
			guard := NewPermissionGuard(plugin.Metadata.Permissions)
			vm := NewJPBCVM(plugin.jpbcModule, guard, r.host)

			// Pass instance as first argument
			methodArgs := append([]interface{}{instance}, args...)

			if _, exists := plugin.jpbcModule.Functions[qualifiedName]; exists {
				return vm.Execute(qualifiedName, methodArgs)
			}
			return vm.Execute(methodName, methodArgs)

		default:
			return nil, fmt.Errorf("formato no soportado para llamadas de metodo: %s", plugin.Format)
		}
	})
}
