package controllers

import (

	//"encoding/json"
	//"fmt"
	"strings"

	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
	ModuleName string
}

func (c *BaseController) Prepare() {
	c.ViewPath = beego.AppConfig.String("viewspath")
	c.Data["viewPath"] = c.ViewPath

	c.Data["baseUrl"] = "/"
	c.Data["resourceUrl"] = "/static/backend/metronic.bootstrap/"
	c.Data["commonResourceUrl"] = "/static/common/"

	controllerName, actionName := c.getControllerAndAction()
	c.Data["controllerName"] = controllerName
	c.Data["actionName"] = actionName
	c.Data["moduleName"] = c.ModuleName
	c.Data["auto_redirect"] = false

}

func (c *BaseController) getControllerAndAction() (string, string) {
	controllerName, actionName := c.GetControllerAndAction()
	controllerName = strings.Replace(controllerName, "Controller", "", 1)
	controllerName = strings.ToLower(controllerName)
	actionName = strings.ToLower(actionName)
	//fmt.Println("controllerName:", controllerName, "actionName:", actionName)
	return controllerName, actionName

}
func (c *BaseController) GetUrl(action string) string {
	controllerName, _ := c.getControllerAndAction()
	//fmt.Println("ModuleName:", c.ModuleName)
	return "/" + strings.Join([]string{c.ModuleName, controllerName, action}, "/")
}
