package main

type ObjectType int

const (
	OBJ_STRING = iota
)

type Obj interface {
	Type() ObjectType
}

type ObjString struct {
	Length     int
	Characters string
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

func AsLiteralString(val Value) string {
	return AsString(val).Characters
}

func IsObjtype(val Value, t ObjectType) bool {
	return isObj(val) && val.AsObj().Type() == t
}

func CreateStringObj(literal string) ObjString {
	return ObjString{Length: len(literal), Characters: literal}
}
