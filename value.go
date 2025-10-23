package main

import "fmt"

// type Value float64

type ValueType int

const (
	VAL_BOOL = iota
	VAL_NIL
	VAL_NUMBER
)

type Value interface {
	Type() ValueType
	AsBoolean() bool
	AsNumber() float64
	Print()
}

type NilVal struct{}

func (nv NilVal) Type() ValueType {
	return VAL_NIL
}

func (nv NilVal) AsBoolean() bool {
	panic("nil value is not a boolean!")
}

func (nv NilVal) AsNumber() float64 {
	panic("nil value is not a number!")
}

func (nv NilVal) Print() {
	fmt.Printf("nil")
}

type BoolVal bool

func (bv BoolVal) Type() ValueType {
	return VAL_BOOL
}

func (bv BoolVal) AsBoolean() bool {
	return bool(bv)
}

func (bv BoolVal) AsNumber() float64 {
	panic("bool value is not a number!")
}

func (bv BoolVal) Print() {
	fmt.Printf("%t", bool(bv))
}

type NumberVal float64

func (nv NumberVal) Type() ValueType {
	return VAL_NUMBER
}

func (nv NumberVal) AsBoolean() bool {
	panic("number value is not a boolean!")
}

func (nv NumberVal) AsNumber() float64 {
	return float64(nv)
}

func (nv NumberVal) Print() {
	fmt.Printf("%g", float64(nv))
}

func isBool(v Value) bool {
	return v.Type() == VAL_BOOL
}

func isNumber(v Value) bool {
	return v.Type() == VAL_NUMBER
}

func isNil(v Value) bool {
	return v.Type() == VAL_NIL
}

func valuesEqual(a, b Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch a.Type() {
	case VAL_BOOL:
		return a.AsBoolean() == b.AsBoolean()
	case VAL_NUMBER:
		return a.AsNumber() == b.AsNumber()
	case VAL_NIL:
		return true
	}

	return false
}

type ValueArray struct {
	values []Value
}
