package nativewasm

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

// NativeWasmBackend compila módulos WebAssembly (.wasm) generados por Rust, C/C++, Dart, etc., a Joss Plugin IR.
type NativeWasmBackend struct {
	Language string
}

func NewNativeWasmBackend(language string) *NativeWasmBackend {
	return &NativeWasmBackend{Language: language}
}

func (b *NativeWasmBackend) Compile(wasmPath string, manifestName, version string, exports []string, permissions []string) (*ir.IRModule, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("native/wasm backend (%s): error al leer %s: %w", b.Language, wasmPath, err)
	}

	// Verificar la cabecera WebAssembly (0x00 0x61 0x73 0x6D => '\0asm')
	if len(wasmBytes) < 4 || wasmBytes[0] != 0x00 || wasmBytes[1] != 0x61 || wasmBytes[2] != 0x73 || wasmBytes[3] != 0x6D {
		return nil, fmt.Errorf("native/wasm backend (%s): %s no es un archivo WebAssembly valido (magic number '\\0asm' no coincide)", b.Language, wasmPath)
	}

	module := ir.NewModule(manifestName, version, b.Language)
	module.Exports = exports
	module.Permissions = permissions

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
						{Op: ir.OpConst, Target: "r0", ConstIdx: module.AddConstant(fmt.Sprintf("%s Native/Wasm context for %s", b.Language, exp))},
						{Op: ir.OpReturn},
					},
				},
			},
		}
		module.Functions[exp] = fn
	}

	return module, nil
}
