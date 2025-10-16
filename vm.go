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
	stack []Value
}

func (vm *VM) Interpret(c *Chunk) InterpretResult {
	vm.chunk = c
	vm.ip = 0
	vm.resetStack()
	return vm.run()
}

func (vm *VM) run() InterpretResult {
	for {
		if DEBUG_TRACE_EXECUTION {
			fmt.Print("          ")
			for _, slot := range vm.stack {
				fmt.Print("[ ")
				fmt.Print(slot)
				fmt.Print(" ]")
			}
			fmt.Println()
			disassembleInstruction(vm.chunk, vm.ip)
		}
		inst := vm.readByte()
		switch inst {
		case OP_RETURN:
			fmt.Print(vm.popStack())
			fmt.Println()
			return INTERPRET_OK
		case OP_CONSTANT:
			constant := vm.readConstant()
			vm.pushStack(constant)
		case OP_NEGATE:
			vm.pushStack(-vm.popStack())
		case OP_ADD:
			vm.performBinaryOp(inst)
		case OP_DIVIDE:
			vm.performBinaryOp(inst)
		case OP_MULTIPLY:
			vm.performBinaryOp(inst)
		case OP_SUBSTRACT:
			vm.performBinaryOp(inst)
		}
	}
}

func (vm *VM) performBinaryOp(operation byte) {
	b := vm.popStack()
	a := vm.popStack()
	switch operation {
	case OP_ADD:
		vm.pushStack(a + b)
	case OP_DIVIDE:
		vm.pushStack(a / b)
	case OP_MULTIPLY:
		vm.pushStack(a * b)
	case OP_SUBSTRACT:
		vm.pushStack(a / b)
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

func (vm *VM) resetStack() {
	vm.stack = []Value{}
}

func (vm *VM) pushStack(val Value) {
	vm.stack = append(vm.stack, val)
}

func (vm *VM) popStack() Value {
	stackLen := len(vm.stack)
	val := vm.stack[stackLen-1]
	vm.stack = vm.stack[:stackLen-1]

	return val
}
