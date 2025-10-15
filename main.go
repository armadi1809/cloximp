package main

func main() {
	chunk := &Chunk{}
	chunk.WriteOpCode(OP_RETURN, 123)
	DisassembleChunk(chunk, "test chunk")

}
