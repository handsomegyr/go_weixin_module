package controllers

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	// 设置cookie

	secret := "index_"
	c.SetSecureCookie(secret, secret+"_cookie", "cookie123", 5400, "/")
	c.Ctx.SetCookie("cookie1", "test1", 3600, "/")
	c.Ctx.SetSecureCookie(secret, secret+"_cookie3", "cookie123", 5400, "/")
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}
