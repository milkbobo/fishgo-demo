package config

import (
	. "github.com/milkbobo/fishgoweb/language"
)

var MethodsMinerEnum struct {
	EnumStruct
	Send int64 `enum:"1,Send"`
}

func init() {
	InitEnumStruct(&MethodsMinerEnum)
}
