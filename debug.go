package main

import "fmt"

func DisassembleChunk(c *Chunk, name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < c.Count(); {
		offset = disassembleInstruction(c, offset)
	}
}

func disassembleInstruction(c *Chunk, offset int) int {
	fmt.Printf("%04d ", offset)
	inst := c.Code[offset]

	switch inst {
	case byte(OP_RETURN):
		return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", inst)
		return offset + 1
	}

}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}
