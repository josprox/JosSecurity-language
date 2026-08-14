package pluginruntime

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

// JPBCInstruction representa una instruccion binaria decodificada.
type JPBCInstruction struct {
	Op       ir.OpCode
	ConstIdx int
}

// JPBCFunction representa una funcion compilada en JPBC.
type JPBCFunction struct {
	Name         string
	IsExported   bool
	Instructions []JPBCInstruction
}

// JPBCModule representa el modulo binario completo decodificado de main.jbc.
type JPBCModule struct {
	MajorVersion uint16
	MinorVersion uint16
	ConstantPool []interface{}
	Functions    map[string]*JPBCFunction
}

// DecodeJPBC deserializa el archivo binario bytecode/main.jbc en formato JPBC.
func DecodeJPBC(data []byte) (*JPBCModule, error) {
	if len(data) < 8 {
		return nil, ErrBytecodeTruncated
	}

	// 1. Validar Magic 'JPBC'
	if string(data[:4]) != "JPBC" {
		return nil, ErrInvalidBytecodeHeader
	}

	reader := bytes.NewReader(data[4:])

	// 2. Validar Version
	var major, minor uint16
	if err := binary.Read(reader, binary.BigEndian, &major); err != nil {
		return nil, fmt.Errorf("error leyendo major version: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &minor); err != nil {
		return nil, fmt.Errorf("error leyendo minor version: %w", err)
	}

	if major != 1 {
		return nil, fmt.Errorf("%w: version %d.%d (esperado 1.x)", ErrUnsupportedVersion, major, minor)
	}

	// 3. Leer Constant Pool
	var constPoolLen uint32
	if err := binary.Read(reader, binary.BigEndian, &constPoolLen); err != nil {
		return nil, fmt.Errorf("error leyendo longitud de constant pool: %w", err)
	}

	constPoolBytes := make([]byte, constPoolLen)
	if _, err := reader.Read(constPoolBytes); err != nil {
		return nil, fmt.Errorf("error leyendo bytes de constant pool: %w", err)
	}

	var constantPool []interface{}
	if len(constPoolBytes) > 0 {
		if err := json.Unmarshal(constPoolBytes, &constantPool); err != nil {
			return nil, fmt.Errorf("error deserializando constant pool JSON: %w", err)
		}
	}

	// 4. Leer Funciones
	var funcCount uint32
	if err := binary.Read(reader, binary.BigEndian, &funcCount); err != nil {
		return nil, fmt.Errorf("error leyendo cantidad de funciones: %w", err)
	}

	functions := make(map[string]*JPBCFunction, funcCount)

	for i := uint32(0); i < funcCount; i++ {
		var nameLen uint16
		if err := binary.Read(reader, binary.BigEndian, &nameLen); err != nil {
			return nil, fmt.Errorf("error leyendo longitud de nombre de funcion %d: %w", i, err)
		}

		nameBytes := make([]byte, nameLen)
		if _, err := reader.Read(nameBytes); err != nil {
			return nil, fmt.Errorf("error leyendo nombre de funcion %d: %w", i, err)
		}
		fnName := string(nameBytes)

		var expFlag uint8
		if err := binary.Read(reader, binary.BigEndian, &expFlag); err != nil {
			return nil, fmt.Errorf("error leyendo export flag de funcion %s: %w", fnName, err)
		}

		var instCount uint32
		if err := binary.Read(reader, binary.BigEndian, &instCount); err != nil {
			return nil, fmt.Errorf("error leyendo cantidad de instrucciones de funcion %s: %w", fnName, err)
		}

		instructions := make([]JPBCInstruction, instCount)
		for j := uint32(0); j < instCount; j++ {
			var op uint8
			var constIdx int32
			if err := binary.Read(reader, binary.BigEndian, &op); err != nil {
				return nil, fmt.Errorf("error leyendo op de instruccion %d en %s: %w", j, fnName, err)
			}
			if err := binary.Read(reader, binary.BigEndian, &constIdx); err != nil {
				return nil, fmt.Errorf("error leyendo constIdx de instruccion %d en %s: %w", j, fnName, err)
			}

			// Validar limites del constant pool si la instruccion hace referencia
			if constIdx >= 0 && int(constIdx) >= len(constantPool) {
				return nil, fmt.Errorf("instruccion %d en %s referencia constant pool fuera de limites (%d >= %d)",
					j, fnName, constIdx, len(constantPool))
			}

			instructions[j] = JPBCInstruction{
				Op:       ir.OpCode(op),
				ConstIdx: int(constIdx),
			}
		}

		functions[fnName] = &JPBCFunction{
			Name:         fnName,
			IsExported:   expFlag == 1,
			Instructions: instructions,
		}
	}

	return &JPBCModule{
		MajorVersion: major,
		MinorVersion: minor,
		ConstantPool: constantPool,
		Functions:    functions,
	}, nil
}
