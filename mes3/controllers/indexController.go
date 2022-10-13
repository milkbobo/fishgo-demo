package controllers

import (
	. "mes3/models/config"
)

type IndexController struct {
	BaseController
	ConfigAo ConfigAoModel
}

func (this *IndexController) Test_Json() interface{} {
	return this.ConfigAo.Get("test")
}
