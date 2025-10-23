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
	if offset > 0 && c.lines[offset] == c.lines[offset-1] {
		fmt.Print("   | ")
	} else {
		fmt.Printf("%4d ", c.lines[offset])
	}

	switch inst {
	case OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	case OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", offset, c)
	case OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case OP_SUBSTRACT:
		return simpleInstruction("OP_SUBSTRACT", offset)
	case OP_NIL:
		return simpleInstruction("OP_NIL", offset)
	case OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", inst)
		return offset + 1
	}

}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func constantInstruction(name string, offset int, c *Chunk) int {
	constantIdx := c.Code[offset+1]
	fmt.Printf("%-16s %4d '", name, constantIdx)
	fmt.Print(c.Constants.values[constantIdx])

	fmt.Printf("'\n")
	return offset + 2
}
