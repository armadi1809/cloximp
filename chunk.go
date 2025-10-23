package main

const (
	OP_RETURN byte = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
	OP_SUBSTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NIL
	OP_TRUE
	OP_FALSE
)

type Chunk struct {
	Code      []byte
	Constants ValueArray
	lines     []int
}

func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	c.lines = append(c.lines, line)
}

func (c *Chunk) Count() int {
	return len(c.Code)
}

func (c *Chunk) AddConstant(val Value) int {
	c.Constants.values = append(c.Constants.values, val)
	return len(c.Constants.values) - 1
}
