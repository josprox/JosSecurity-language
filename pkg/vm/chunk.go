package vm

type Chunk struct {
	Code      []byte
	Constants []Value
	Lines     []int
}

func NewChunk() *Chunk {
	return &Chunk{
		Code:      make([]byte, 0, 64),
		Constants: make([]Value, 0, 16),
		Lines:     make([]int, 0, 64),
	}
}

func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	c.Lines = append(c.Lines, line)
}

func (c *Chunk) WriteOp(op Opcode, line int) {
	c.Write(byte(op), line)
}

func (c *Chunk) AddConstant(val Value) int {
	for i, existing := range c.Constants {
		if existing == val {
			return i
		}
	}
	c.Constants = append(c.Constants, val)
	return len(c.Constants) - 1
}

func (c *Chunk) WriteConst(val Value, line int) {
	idx := c.AddConstant(val)
	c.WriteOp(OpConst, line)
	c.Write(byte(idx), line)
}
