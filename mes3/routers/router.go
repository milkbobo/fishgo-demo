package routers

import (
	. "github.com/milkbobo/fishgoweb/web"
	. "mes3/controllers"
)

func init() {
	//前端路由
	InitRoute("/index", &IndexController{})
}
