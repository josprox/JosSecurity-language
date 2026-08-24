package pluginruntime

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidBytecodeHeader = errors.New("pluginruntime: cabecera magica de bytecode no reconocida (esperado JOSSBC2Z o JPBC)")
	ErrBytecodeTruncated     = errors.New("pluginruntime: archivo de bytecode truncado o incompleto")
	ErrUnsupportedVersion    = errors.New("pluginruntime: version de formato JPBC no soportada")
	ErrPluginNotFound        = errors.New("pluginruntime: plugin no encontrado en el registro")
	ErrFunctionNotFound      = errors.New("pluginruntime: funcion no encontrada en el plugin")
	ErrPluginAlreadyLoaded   = errors.New("pluginruntime: plugin ya registrado con el mismo nombre y version")
	ErrExecutionLimitExceeded = errors.New("pluginruntime: limite de instrucciones excedido (posible bucle infinito)")
)

// PluginError encapsula fallos ocurridos durante la ejecucion de un plugin con contexto.
type PluginError struct {
	Plugin   string
	Function string
	PC       int
	Op       string
	Cause    error
}

func (e *PluginError) Error() string {
	if e.Function != "" {
		return fmt.Sprintf("error en plugin %s::%s() [PC: %d, Op: %s]: %v", e.Plugin, e.Function, e.PC, e.Op, e.Cause)
	}
	return fmt.Sprintf("error en plugin %s: %v", e.Plugin, e.Cause)
}

func (e *PluginError) Unwrap() error {
	return e.Cause
}

// SafeCall ejecuta una funcion dentro de una frontera de aislamiento y captura panics.
func SafeCall(pluginName, fnName string, fn func() (interface{}, error)) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &PluginError{
				Plugin:   pluginName,
				Function: fnName,
				Cause:    fmt.Errorf("panic atrapado en tiempo de ejecucion: %v", r),
			}
		}
	}()
	return fn()
}
