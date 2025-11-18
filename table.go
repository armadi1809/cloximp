package main

import "maps"

type Table map[string]Value

func InitTable() Table {
	return Table(make(map[string]Value))
}

func (tb Table) TableSet(key ObjString, val Value) bool {
	entry := key.Characters
	_, existed := tb[entry]
	tb[entry] = val
	return existed
}

func (tb Table) TableGet(key ObjString) (Value, bool) {
	entry := key.Characters
	if v, ok := tb[entry]; ok {
		return v, ok
	}
	return nil, false
}

func TableAddAll(from, to Table) {
	maps.Copy(to, from)
}

func (tb Table) TableDelete(key ObjString) bool {
	entry := key.Characters
	_, existed := tb[entry]
	delete(tb, entry)
	return existed
}
