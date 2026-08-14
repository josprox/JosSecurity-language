package optimizer

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

// Result contiene el modulo optimizado y estadisticas de eliminacion de codigo.
type Result struct {
	Module           *ir.IRModule
	OriginalFuncs    int
	OptimizedFuncs   int
	RemovedFuncs     int
	OriginalStructs  int
	OptimizedStructs int
	RemovedStructs   int
}

// TreeShake realiza el analisis de alcanzabilidad desde los metodos exportados
// y elimina las funciones, estructuras y campos no utilizados de dependencias externas.
func TreeShake(module *ir.IRModule) (*Result, error) {
	if module == nil {
		return nil, fmt.Errorf("tree shaker: modulo es nil")
	}
	if len(module.Exports) == 0 {
		return nil, fmt.Errorf("tree shaker: el modulo no declara ninguna funcion exportada")
	}

	reachableFuncs := make(map[string]bool)
	reachableStructs := make(map[string]bool)

	// Inicializar worklist con todas las funciones exportadas del modulo
	worklist := make([]string, 0)
	for _, exp := range module.Exports {
		if _, ok := module.Functions[exp]; ok {
			if !reachableFuncs[exp] {
				reachableFuncs[exp] = true
				worklist = append(worklist, exp)
			}
		}
	}
	for name, fn := range module.Functions {
		if fn != nil && fn.IsExported {
			if !reachableFuncs[name] {
				reachableFuncs[name] = true
				worklist = append(worklist, name)
			}
		}
	}

	// Recorrer el grafo de llamadas (Call Graph Reachability) de forma iterativa y segura contra ciclos
	for len(worklist) > 0 {
		curr := worklist[0]
		worklist = worklist[1:]

		fn, ok := module.Functions[curr]
		if !ok || fn == nil {
			continue
		}

		// Inspeccionar instrucciones de la funcion alcanzable
		for _, block := range fn.Blocks {
			if block == nil {
				continue
			}
			for _, inst := range block.Instructions {
				switch inst.Op {
				case ir.OpCallStatic, ir.OpCallVirtual:
					if len(inst.Args) > 0 {
						targetFn := inst.Args[0]
						if targetFn != "" && !reachableFuncs[targetFn] {
							reachableFuncs[targetFn] = true
							worklist = append(worklist, targetFn)
						}
					}
				case ir.OpNewObject, ir.OpGetField, ir.OpSetField:
					if len(inst.Args) > 0 {
						structName := inst.Args[0]
						if structName != "" {
							reachableStructs[structName] = true
						}
					}
				}
			}
		}
	}

	res := &Result{
		OriginalFuncs:   len(module.Functions),
		OriginalStructs: len(module.Structs),
	}

	// Filtrar funciones: conservar exclusivamente las alcanzables
	optimizedFunctions := make(map[string]*ir.IRFunction)
	for name, fn := range module.Functions {
		if reachableFuncs[name] {
			optimizedFunctions[name] = fn
		}
	}
	module.Functions = optimizedFunctions

	// Filtrar estructuras: conservar exclusivamente las alcanzables
	optimizedStructs := make(map[string]*ir.IRStruct)
	for name, str := range module.Structs {
		if reachableStructs[name] {
			optimizedStructs[name] = str
		}
	}
	module.Structs = optimizedStructs

	res.Module = module
	res.OptimizedFuncs = len(module.Functions)
	res.RemovedFuncs = res.OriginalFuncs - res.OptimizedFuncs
	res.OptimizedStructs = len(module.Structs)
	res.RemovedStructs = res.OriginalStructs - res.OptimizedStructs

	return res, nil
}
