package main

type OpCode byte

const (
	OP_RETURN OpCode = iota
	OP_CONSTANT
)

type Chunk struct {
	Code      []byte
	Constants []int
	lines     []int
}

func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	c.lines = append(c.lines, line)
}

func (c *Chunk) Count() int {
	return len(c.Code)
}

func (c *Chunk) WriteOpCode(o OpCode, line int) {
	c.Code = append(c.Code, byte(o))
	c.lines = append(c.lines, line)
}
