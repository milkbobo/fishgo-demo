package routers

import (
	. "github.com/milkbobo/fishgoweb/web"
	. "sell999/controllers"

)

func init() {
	//前端路由
	InitRoute("/index", &IndexController{})
}
