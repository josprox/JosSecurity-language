package php

import (
	"fmt"
	"os"
	"strings"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

// Backend para compilar código PHP a Joss Plugin IR.
type PHPBackend struct{}

func NewPHPBackend() *PHPBackend {
	return &PHPBackend{}
}

func (b *PHPBackend) Compile(sourcePath string, manifestName, version string, exports []string, permissions []string) (*ir.IRModule, error) {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("php backend: error al leer %s: %w", sourcePath, err)
	}

	module := ir.NewModule(manifestName, version, "php")
	module.Exports = exports
	module.Permissions = permissions

	code := string(data)
	lines := strings.Split(code, "\n")
	definedFuncs := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "function ") {
			parts := strings.Split(trimmed[9:], "(")
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
			ReturnType: "dynamic",
			IsExported: true,
			Blocks: []*ir.IRBlock{
				{
					Label: "entry",
					Instructions: []ir.IRInstruction{
						{Op: ir.OpConst, Target: "r0", ConstIdx: module.AddConstant(fmt.Sprintf("PHP execution context for %s", exp))},
						{Op: ir.OpReturn},
					},
				},
			},
		}
		module.Functions[exp] = fn
	}

	for fnName := range definedFuncs {
		if _, exists := module.Functions[fnName]; !exists {
			module.Functions[fnName] = &ir.IRFunction{
				Name:       fnName,
				Params:     make([]ir.IRField, 0),
				ReturnType: "dynamic",
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

	return module, nil
}
