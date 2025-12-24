package main

import (
	"fmt"
	"os"
)

type InterpretResult byte

const DEBUG_TRACE_EXECUTION = true
const FRAME_MAX = 64

const (
	INTERPRET_OK = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type CallFrame struct {
	function *ObjFunction
	ip       int
	slots    int
}

type VM struct {
	frames     []CallFrame
	frameCount int
	stack      []Value
	compiler   *Compiler
	globals    map[ObjString]Value
}

func (vm *VM) initVM() {
	vm.initCompiler(TYPE_SCRIPT)
	vm.resetStack()
	vm.globals = make(map[ObjString]Value)
}

func (vm *VM) Interpret(source string) InterpretResult {
	vm.initVM()
	function := vm.compiler.compile(source)
	if function == nil {
		return INTERPRET_COMPILE_ERROR
	}
	vm.pushStack(ObjVal{Object: function})
	frame := CallFrame{
		function: function,
		ip:       0,
		slots:    len(vm.stack) - 1,
	}
	vm.frames = append(vm.frames, frame)

	return vm.run()
}

func (vm *VM) initCompiler(funct FunctionType) {
	local := Local{
		depth: 0,
		name: Token{
			Lexeme: "",
		},
	}
	vm.compiler = &Compiler{
		Sc:       &Scanner{},
		Function: NewFunction(),
		Type:     funct,
		Ps: &Parser{
			panicMode: false,
			hadError:  false,
		},
		Locals:     []Local{local},
		ScopeDepth: 0,
		LocalCount: 1,
	}
}
func (vm *VM) run() InterpretResult {
	frame := vm.getCurrentFrame()
	for {
		if DEBUG_TRACE_EXECUTION {
			fmt.Print("          ")
			for _, slot := range vm.stack {
				fmt.Print("[ ")
				slot.Print()
				fmt.Print(" ]")
			}
			fmt.Println()
			disassembleInstruction(&frame.function.chunk, frame.ip)
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
		case OP_POP:
			vm.popStack()
		case OP_DEFINE_GLOBAL:
			name := vm.readString()
			vm.globals[name] = vm.peek(0)
			vm.popStack()
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
		case OP_PRINT:
			vm.popStack().Print()
			fmt.Print("\n")
		case OP_GET_GLOBAL:
			name := vm.readString()
			val, ok := vm.globals[name]
			if !ok {
				vm.runtimeError("Undefined variable '%s'.", name.Characters)
				return INTERPRET_RUNTIME_ERROR
			}
			vm.pushStack(val)
		case OP_GET_LOCAL:
			slot := vm.readByte()
			vm.pushStack(vm.stack[frame.slots+int(slot)])
		case OP_SET_LOCAL:
			slot := vm.readByte()
			vm.stack[frame.slots+int(slot)] = vm.peek(0)
		case OP_SET_GLOBAL:
			name := vm.readString()
			_, ok := vm.globals[name]
			if !ok {
				vm.runtimeError("Undefined variable '%s'.", name.Characters)
				return INTERPRET_RUNTIME_ERROR
			}
			vm.globals[name] = vm.peek(0)
		case OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if isFalsey(vm.peek(0)) {
				frame.ip += offset
			}
		case OP_JUMP:
			offset := vm.readShort()
			frame.ip += offset
		case OP_LOOP:
			offset := vm.readShort()
			frame.ip -= offset
		}
	}
}

func (vm *VM) readShort() int {
	frame := vm.getCurrentFrame()
	frame.ip += 2
	high := uint16(frame.function.chunk.Code[frame.ip-2])
	low := uint16(frame.function.chunk.Code[frame.ip-1])

	return int((high << 8) | low)
}

func (vm *VM) performBinaryOp(operation byte) bool {

	switch operation {
	case OP_ADD:
		if IsString(vm.peek(0)) && IsString(vm.peek(1)) {
			b := AsString(vm.popStack())
			a := AsString(vm.popStack())
			vm.pushStack(ObjVal{Object: CreateStringObj(a.Characters + b.Characters)})
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
	frame := vm.getCurrentFrame()
	inst := frame.function.chunk.Code[frame.ip]
	frame.ip += 1
	return inst
}

func (vm *VM) readConstant() Value {
	frame := vm.getCurrentFrame()
	return frame.function.chunk.Constants.values[vm.readByte()]
}

func (vm *VM) readString() ObjString {
	return AsString(vm.readConstant())
}

func (vm *VM) resetStack() {
	vm.stack = []Value{}
	vm.frameCount = 0
	vm.frames = []CallFrame{}
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
	frame := vm.getCurrentFrame()
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	instruction := frame.ip - 1
	line := frame.function.chunk.lines[instruction]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
}

func (vm *VM) getCurrentFrame() *CallFrame {
	return &vm.frames[vm.frameCount-1]
}
