package main

import "fmt"

type Value float64

type ValueArray struct {
	values []Value
}

func (v Value) Print() {
	fmt.Printf("%g", v)
}
