package routers

import (
	"go_weixin_module/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	beego.Router("/weixin/index/index", &controllers.MainController{}, "get:Get")
	beego.Router("/weixin/sns/index", &controllers.SnsController{}, "get:Index")
	beego.Router("/weixin/sns/callback", &controllers.SnsController{}, "get:Callback")

}
