package pluginruntime

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/bytecode"
	"github.com/jossecurity/joss/pkg/parser"
)

// JossASTExecutor ejecuta plugins nativos de Joss basados en AST (JOSSBC2Z).
type JossASTExecutor struct {
	Program *parser.Program
	Engine  ASTEngine
}

// NewJossASTExecutor decodifica el bytecode AST y prepara el ejecutor.
func NewJossASTExecutor(bytecodeBytes []byte, engine ASTEngine) (*JossASTExecutor, error) {
	program, err := bytecode.Decode(bytecodeBytes)
	if err != nil {
		return nil, fmt.Errorf("joss_executor: error decodificando AST JOSSBC2Z: %w", err)
	}

	if engine != nil {
		if err := engine.RegisterProgram(program); err != nil {
			return nil, err
		}
	}

	return &JossASTExecutor{
		Program: program,
		Engine:  engine,
	}, nil
}

// CallFunction invoca una funcion global exportada por el plugin Joss a traves del motor AST.
func (e *JossASTExecutor) CallFunction(fnName string, args []interface{}) (interface{}, error) {
	if e.Engine == nil {
		return nil, fmt.Errorf("joss_executor: motor AST no inicializado")
	}
	return e.Engine.CallFunction(fnName, args)
}

// Instantiate crea una nueva instancia de una clase exportada por el plugin.
func (e *JossASTExecutor) Instantiate(className string, args []interface{}) (interface{}, error) {
	if e.Engine == nil {
		return nil, fmt.Errorf("joss_executor: motor AST no inicializado")
	}
	return e.Engine.Instantiate(className, args)
}

// CallMethod invoca un metodo de instancia sobre una clase del plugin.
func (e *JossASTExecutor) CallMethod(instance interface{}, methodName string, args []interface{}) (interface{}, error) {
	if e.Engine == nil {
		return nil, fmt.Errorf("joss_executor: motor AST no inicializado")
	}
	return e.Engine.CallMethod(instance, methodName, args)
}
