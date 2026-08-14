package codegen

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
	"github.com/jossecurity/joss/pkg/pluginpkg"
)

// CodeGenerator convierte un IRModule optimizado a Bytecode binario de Joss (JPBC).
type CodeGenerator struct{}

// NewCodeGenerator crea una instancia de CodeGenerator.
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

// GenerateBytecode produce los bytes del archivo main.jbc y el indice de simbolos desde Joss IR.
func (cg *CodeGenerator) GenerateBytecode(module *ir.IRModule) ([]byte, []byte, error) {
	buf := new(bytes.Buffer)

	// Magic header para JPBC (Joss Plugin ByteCode)
	magic := []byte{'J', 'P', 'B', 'C'}
	buf.Write(magic)

	// Escribir version del bytecode (1.0)
	binary.Write(buf, binary.BigEndian, uint16(1))
	binary.Write(buf, binary.BigEndian, uint16(0))

	// Escribir metadatos de exportaciones y pool de constantes
	constPoolJSON, err := json.Marshal(module.ConstantPool)
	if err != nil {
		return nil, nil, fmt.Errorf("codegen: error al codificar piscina de constantes: %w", err)
	}

	binary.Write(buf, binary.BigEndian, uint32(len(constPoolJSON)))
	buf.Write(constPoolJSON)

	// Escribir cantidad de funciones
	binary.Write(buf, binary.BigEndian, uint32(len(module.Functions)))

	// Sort function names for deterministic output
	fnNames := make([]string, 0, len(module.Functions))
	for name := range module.Functions {
		fnNames = append(fnNames, name)
	}
	sort.Strings(fnNames)

	var classes []pluginpkg.SymbolClass
	var functions []pluginpkg.SymbolCallable

	structNames := make([]string, 0, len(module.Structs))
	for name := range module.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	for _, name := range structNames {
		str := module.Structs[name]
		var props []string
		for _, f := range str.Fields {
			props = append(props, f.Name)
		}
		classes = append(classes, pluginpkg.SymbolClass{
			Name:       name,
			Properties: props,
		})
	}

	for _, name := range fnNames {
		fn := module.Functions[name]

		// Nombre de funcion
		fnNameBytes := []byte(name)
		binary.Write(buf, binary.BigEndian, uint16(len(fnNameBytes)))
		buf.Write(fnNameBytes)

		// Flag Exported
		var expFlag uint8 = 0
		if fn.IsExported {
			expFlag = 1
			var params []pluginpkg.SymbolParameter
			for _, p := range fn.Params {
				params = append(params, pluginpkg.SymbolParameter{
					Name: p.Name,
					Type: p.Type,
				})
			}
			functions = append(functions, pluginpkg.SymbolCallable{
				Name:       name,
				Parameters: params,
			})
		}
		binary.Write(buf, binary.BigEndian, expFlag)

		// Escribir bloques e instrucciones
		totalInsts := 0
		for _, blk := range fn.Blocks {
			totalInsts += len(blk.Instructions)
		}
		binary.Write(buf, binary.BigEndian, uint32(totalInsts))

		for _, blk := range fn.Blocks {
			for _, inst := range blk.Instructions {
				binary.Write(buf, binary.BigEndian, uint8(inst.Op))
				binary.Write(buf, binary.BigEndian, int32(inst.ConstIdx))
			}
		}
	}

	symbolIndex := pluginpkg.BuildSymbolIndexFromCallables(module.Name, module.Version, classes, functions)
	symbolsJSON, err := json.MarshalIndent(symbolIndex, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("codegen: error al generar indice de simbolos: %w", err)
	}

	return buf.Bytes(), symbolsJSON, nil
}
