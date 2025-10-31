package main

type ObjectType int

const (
	OBJ_STRING = iota
)

type Obj struct {
	ObType ObjectType
}

type ObjString struct {
	Obj
	Length     int
	Characters string
}
