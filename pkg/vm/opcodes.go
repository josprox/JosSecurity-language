package vm

type Opcode uint8

const (
	OpHalt Opcode = iota
	OpConst
	OpLoadLocal
	OpStoreLocal
	OpAddInt
	OpSubInt
	OpMulInt
	OpDivFloat
	OpModInt
	OpNegInt
	OpCmpLt
	OpCmpLe
	OpCmpGt
	OpCmpGe
	OpCmpEq
	OpCmpNe
	OpJump
	OpJumpIfFalse
	OpJumpIfTrue
	OpCall
	OpReturn
	OpPop
)

func (op Opcode) String() string {
	switch op {
	case OpHalt:
		return "OP_HALT"
	case OpConst:
		return "OP_CONST"
	case OpLoadLocal:
		return "OP_LOAD_LOCAL"
	case OpStoreLocal:
		return "OP_STORE_LOCAL"
	case OpAddInt:
		return "OP_ADD_INT"
	case OpSubInt:
		return "OP_SUB_INT"
	case OpMulInt:
		return "OP_MUL_INT"
	case OpDivFloat:
		return "OP_DIV_FLOAT"
	case OpModInt:
		return "OP_MOD_INT"
	case OpNegInt:
		return "OP_NEG_INT"
	case OpCmpLt:
		return "OP_CMP_LT"
	case OpCmpLe:
		return "OP_CMP_LE"
	case OpCmpGt:
		return "OP_CMP_GT"
	case OpCmpGe:
		return "OP_CMP_GE"
	case OpCmpEq:
		return "OP_CMP_EQ"
	case OpCmpNe:
		return "OP_CMP_NE"
	case OpJump:
		return "OP_JUMP"
	case OpJumpIfFalse:
		return "OP_JUMP_IF_FALSE"
	case OpJumpIfTrue:
		return "OP_JUMP_IF_TRUE"
	case OpCall:
		return "OP_CALL"
	case OpReturn:
		return "OP_RETURN"
	case OpPop:
		return "OP_POP"
	default:
		return "OP_UNKNOWN"
	}
}
