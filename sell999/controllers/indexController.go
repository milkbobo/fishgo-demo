package controllers

import (
	. "sell999/models/config"
)

type IndexController struct {
	BaseController
	ConfigAo ConfigAoModel
}

func (this *IndexController) Test_Json() interface{} {
	return this.ConfigAo.Get("test")
}
