package main

import "fmt"

type InterpretResult byte

const DEBUG_TRACE_EXECUTION = true

const (
	INTERPRET_OK = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	chunk *Chunk
	ip    int
}

func (vm *VM) Interpret(c *Chunk) InterpretResult {
	vm.chunk = c
	vm.ip = 0

	return vm.run()
}

func (vm *VM) run() InterpretResult {
	for {
		if DEBUG_TRACE_EXECUTION {
			disassembleInstruction(vm.chunk, vm.ip)
		}
		inst := vm.readByte()
		switch inst {
		case OP_RETURN:
			return INTERPRET_OK
		case OP_CONSTANT:
			constant := vm.readConstant()
			fmt.Print(constant)
			fmt.Println()
		}
	}
}

func (vm *VM) readByte() byte {
	inst := vm.chunk.Code[vm.ip]
	vm.ip += 1
	return inst
}

func (vm *VM) readConstant() Value {
	return vm.chunk.Constants.values[vm.readByte()]
}
