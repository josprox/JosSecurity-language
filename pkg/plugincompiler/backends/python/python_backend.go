package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

// Backend para compilar código o scripts Python a Joss Plugin IR.
type PythonBackend struct{}

func NewPythonBackend() *PythonBackend {
	return &PythonBackend{}
}

func (b *PythonBackend) Compile(sourcePath string, manifestName, version string, exports []string, permissions []string) (*ir.IRModule, error) {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("python backend: error al leer %s: %w", sourcePath, err)
	}

	module := ir.NewModule(manifestName, version, "python")
	module.Exports = exports
	module.Permissions = permissions

	code := string(data)

	// Extraer funciones 'def' definidas en el script Python
	lines := strings.Split(code, "\n")
	definedFuncs := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "def ") {
			parts := strings.Split(trimmed[4:], "(")
			if len(parts) > 0 {
				funcName := strings.TrimSpace(parts[0])
				definedFuncs[funcName] = true
			}
		}
	}

	for _, exp := range exports {
		fn := &ir.IRFunction{
			Name:       exp,
			Params:     make([]ir.IRField, 0),
			ReturnType: "mixed",
			IsExported: true,
			Blocks: []*ir.IRBlock{
				{
					Label: "entry",
					Instructions: []ir.IRInstruction{
						{Op: ir.OpConst, Target: "r0", ConstIdx: module.AddConstant(fmt.Sprintf("Python execution context for %s", exp))},
						{Op: ir.OpReturn},
					},
				},
			},
		}
		module.Functions[exp] = fn
	}

	// Registrar funciones secundarias detectadas en Python para el Tree Shaker
	for fnName := range definedFuncs {
		if _, exists := module.Functions[fnName]; !exists {
			module.Functions[fnName] = &ir.IRFunction{
				Name:       fnName,
				Params:     make([]ir.IRField, 0),
				ReturnType: "mixed",
				IsExported: false,
				Blocks: []*ir.IRBlock{
					{
						Label: "entry",
						Instructions: []ir.IRInstruction{
							{Op: ir.OpReturn},
						},
					},
				},
			}
		}
	}

	_ = filepath.Base(sourcePath)
	return module, nil
}
