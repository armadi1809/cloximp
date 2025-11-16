package main

import (
	"fmt"
	"os"
)

type InterpretResult byte

const DEBUG_TRACE_EXECUTION = true

const (
	INTERPRET_OK = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	ip       int
	stack    []Value
	compiler *Compiler
}

func (vm *VM) initVM() {
	vm.initCompiler()
	vm.ip = 0
	vm.compiler = &Compiler{}
	vm.resetStack()
}

func (vm *VM) Interpret(source string) InterpretResult {
	vm.initCompiler()

	if !vm.compiler.compile(source) {
		return INTERPRET_COMPILE_ERROR
	}

	return vm.run()
}

func (vm *VM) initCompiler() {
	vm.compiler = &Compiler{
		Sc:    &Scanner{},
		Chunk: &Chunk{},
		Ps: &Parser{
			panicMode: false,
			hadError:  false,
		},
	}
}
func (vm *VM) run() InterpretResult {
	for {
		if DEBUG_TRACE_EXECUTION {
			fmt.Print("          ")
			for _, slot := range vm.stack {
				fmt.Print("[ ")
				slot.Print()
				fmt.Print(" ]")
			}
			fmt.Println()
			disassembleInstruction(vm.compiler.Chunk, vm.ip)
		}
		inst := vm.readByte()
		switch inst {
		case OP_RETURN:
			return INTERPRET_OK
		case OP_CONSTANT:
			constant := vm.readConstant()
			vm.pushStack(constant)
		case OP_NEGATE:
			if !isNumber(vm.peek(0)) {
				vm.runtimeError("Operand must be a number")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.pushStack(NumberVal(-vm.popStack().AsNumber()))
		case OP_NOT:
			vm.pushStack(BoolVal(isFalsey(vm.popStack())))
		case OP_ADD:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}

		case OP_DIVIDE:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_MULTIPLY:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_SUBSTRACT:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_NIL:
			vm.pushStack(NilVal{})
		case OP_TRUE:
			vm.pushStack(BoolVal(true))
		case OP_FALSE:
			vm.pushStack(BoolVal(false))
		case OP_EQUAL:
			vm.pushStack(BoolVal(valuesEqual(vm.popStack(), vm.popStack())))
		case OP_LESS:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}
		case OP_GREATER:
			if !vm.performBinaryOp(inst) {
				return INTERPRET_RUNTIME_ERROR
			}
		}
	}
}

func (vm *VM) performBinaryOp(operation byte) bool {

	switch operation {
	case OP_ADD:
		if IsString(vm.peek(0)) && IsString(vm.peek(1)) {
			b := AsString(vm.popStack())
			a := AsString(vm.popStack().AsObj())
			vm.pushStack(CreateStringObj(a.Characters + b.Characters))
		} else if isNumber(vm.peek(0)) && isNumber(vm.peek(1)) {
			b := vm.popStack().AsNumber()
			a := vm.popStack().AsNumber()
			vm.pushStack(NumberVal(a + b))
		} else {
			vm.runtimeError("Operands must be two numbers or two strings")
			return false
		}
	case OP_DIVIDE:
		b := vm.popStack().AsNumber()
		a := vm.popStack().AsNumber()
		vm.pushStack(NumberVal(a / b))
	case OP_MULTIPLY:
		b := vm.popStack().AsNumber()
		a := vm.popStack().AsNumber()
		vm.pushStack(NumberVal(a * b))
	case OP_SUBSTRACT:
		b := vm.popStack().AsNumber()
		a := vm.popStack().AsNumber()
		vm.pushStack(NumberVal(a - b))
	case OP_GREATER:
		b := vm.popStack().AsNumber()
		a := vm.popStack().AsNumber()
		vm.pushStack(BoolVal(a > b))
	case OP_LESS:
		b := vm.popStack().AsNumber()
		a := vm.popStack().AsNumber()
		vm.pushStack(BoolVal(a < b))
	}
	return true
}

func (vm *VM) readByte() byte {
	inst := vm.compiler.Chunk.Code[vm.ip]
	vm.ip += 1
	return inst
}

func (vm *VM) readConstant() Value {
	return vm.compiler.Chunk.Constants.values[vm.readByte()]
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

func (vm *VM) peek(distance int) Value {
	return vm.stack[len(vm.stack)-1-distance]
}

func isFalsey(v Value) bool {
	return isNil(v) || (isBool(v) && !v.AsBoolean())
}

func (vm *VM) runtimeError(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	instruction := vm.ip - 1
	line := vm.compiler.Chunk.lines[instruction]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
}
