package test

import (
	. "github.com/milkbobo/fishgoweb/web"
)

type BaseAoModel struct {
	Model
	ConfigAo ConfigAoModel
}

type ExtendAoModel struct {
	BaseAoModel
}
