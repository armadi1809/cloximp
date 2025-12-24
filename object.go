package main

type ObjectType int

const (
	OBJ_STRING = iota
	OBJ_FUNCTION
)

type Obj interface {
	Type() ObjectType
}

type ObjString struct {
	Length     int
	Characters string
}

type ObjFunction struct {
	arity int
	chunk Chunk
	name  *ObjString
}

func (ObjFunction) Type() ObjectType {
	return OBJ_FUNCTION
}

func (ObjString) Type() ObjectType {
	return OBJ_STRING
}

func AsString(val Value) ObjString {
	if objStr, ok := val.AsObj().(ObjString); ok {
		return objStr
	}
	panic("value is not a string object")
}

func AsFunc(val Value) ObjFunction {
	if objStr, ok := val.AsObj().(ObjFunction); ok {
		return objStr
	}
	panic("value is not a function object")
}

func AsLiteralString(val Value) string {
	return AsString(val).Characters
}

func IsObjtype(val Value, t ObjectType) bool {
	return isObj(val) && val.AsObj().Type() == t
}

func CreateStringObj(literal string) ObjString {
	return ObjString{Length: len(literal), Characters: literal}
}

func NewFunction() ObjFunction {
	return ObjFunction{
		arity: 0,
		name:  nil,
		chunk: Chunk{},
	}
}
