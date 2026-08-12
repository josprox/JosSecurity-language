package codegen

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
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

	symbolsMap := make(map[string]interface{})

	for name, fn := range module.Functions {
		// Nombre de funcion
		fnNameBytes := []byte(name)
		binary.Write(buf, binary.BigEndian, uint16(len(fnNameBytes)))
		buf.Write(fnNameBytes)

		// Flag Exported
		var expFlag uint8 = 0
		if fn.IsExported {
			expFlag = 1
			symbolsMap[name] = map[string]interface{}{
				"name":        name,
				"exported":    true,
				"return_type": fn.ReturnType,
			}
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

	symbolsJSON, err := json.MarshalIndent(symbolsMap, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("codegen: error al generar indice de simbolos: %w", err)
	}

	return buf.Bytes(), symbolsJSON, nil
}
