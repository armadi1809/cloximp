package main

import "fmt"

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

func AsString(val Value) *ObjString {
	if objStr, ok := val.(*ObjString); ok {
		return objStr
	}
	panic("value is not a string object")
}

func AsLiteralString(val Value) string {
	if objStr, ok := val.(*ObjString); ok {
		return objStr.Characters
	}
	panic("value is not a string object")
}

func IsObjtype(val Value, t ObjectType) bool {
	return isObj(val) && val.AsObj().ObType == t
}

func CreateStringObj(literal string) *ObjString {
	return &ObjString{Obj: CreateObj(OBJ_STRING), Length: len(literal), Characters: literal}
}

func CreateObj(objTyp ObjectType) Obj {
	return Obj{ObType: objTyp}
}

func printObj(ob Obj) {
	switch ob.ObType {
	case OBJ_STRING:
		fmt.Print(AsLiteralString(ob))
	}
}
