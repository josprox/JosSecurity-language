package ir

import "fmt"

// IRModule representa un modulo compilado de plugin en Joss Plugin IR.
type IRModule struct {
	Name         string
	Version      string
	Language     string
	Exports      []string
	Permissions  []string
	Structs      map[string]*IRStruct
	Functions    map[string]*IRFunction
	GlobalVars   map[string]*IRGlobalVar
	ConstantPool []interface{}
}

// NewModule crea un nuevo IRModule inicializado.
func NewModule(name, version, language string) *IRModule {
	return &IRModule{
		Name:         name,
		Version:      version,
		Language:     language,
		Exports:      make([]string, 0),
		Permissions:  make([]string, 0),
		Structs:      make(map[string]*IRStruct),
		Functions:    make(map[string]*IRFunction),
		GlobalVars:   make(map[string]*IRGlobalVar),
		ConstantPool: make([]interface{}, 0),
	}
}

// AddConstant agrega un valor a la piscina de constantes y retorna su indice.
func (m *IRModule) AddConstant(val interface{}) int {
	for idx, c := range m.ConstantPool {
		if c == val {
			return idx
		}
	}
	m.ConstantPool = append(m.ConstantPool, val)
	return len(m.ConstantPool) - 1
}

// IRStruct define una estructura/clase en el IR.
type IRStruct struct {
	Name   string
	Fields []IRField
}

// IRField representa un campo dentro de una estructura IR.
type IRField struct {
	Name string
	Type string
}

// IRGlobalVar define una variable global.
type IRGlobalVar struct {
	Name string
	Type string
}

// IRFunction representa una funcion en el IR.
type IRFunction struct {
	Name       string
	Params     []IRField
	ReturnType string
	IsExported bool
	Blocks     []*IRBlock
}

// IRBlock es un bloque basico de control de flujo.
type IRBlock struct {
	Label        string
	Instructions []IRInstruction
}

// OpCode define la instruccion en Joss IR.
type OpCode int

const (
	OpNop OpCode = iota
	OpConst
	OpLoad
	OpStore
	OpGetField
	OpSetField
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpCallStatic
	OpCallVirtual
	OpReturn
	OpBranch
	OpBranchIf
	OpNewObject
)

func (op OpCode) String() string {
	switch op {
	case OpNop:
		return "NOP"
	case OpConst:
		return "CONST"
	case OpLoad:
		return "LOAD"
	case OpStore:
		return "STORE"
	case OpGetField:
		return "GET_FIELD"
	case OpSetField:
		return "SET_FIELD"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpCallStatic:
		return "CALL_STATIC"
	case OpCallVirtual:
		return "CALL_VIRTUAL"
	case OpReturn:
		return "RETURN"
	case OpBranch:
		return "BRANCH"
	case OpBranchIf:
		return "BRANCH_IF"
	case OpNewObject:
		return "NEW_OBJECT"
	default:
		return fmt.Sprintf("UNKNOWN_OP(%d)", op)
	}
}

// IRInstruction representa una instruccion dentro de un bloque basico.
type IRInstruction struct {
	Op       OpCode
	Target   string
	Args     []string
	ConstIdx int
}
