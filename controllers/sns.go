package controllers

import (
	"go_weixin_module/library"
	"strings"
	//"errors"
	//"fmt"
	//"net"
	//"net/url"
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/silenceper/wechat"
)

type SnsController struct {
	BaseController
}

func (c *SnsController) Prepare() {
	c.ModuleName = "weixin"
	c.BaseController.Prepare()
}

// 授权处理
func (c *SnsController) Index() {

	appid := library.Trim(c.GetString("appid", ""))
	redirect := library.Trim(c.GetString("redirect", ""))
	state := library.Trim(c.GetString("state", library.Uniqid(5)))

	//默认授权方式
	scope := library.Trim(c.GetString("scope", "snsapi_userinfo"))
	//是否检查回调域名
	//dc := library.Trim(c.GetString("dc", "0"))
	//是否刷新
	//refresh := library.Trim(c.GetString("refresh", "0"))

	valid := validation.Validation{}
	valid.Required(appid, "appid Can not be empty.")
	valid.Required(redirect, "redirect Can not be empty.")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			//panic(errors.New(err.Key))
			c.Ctx.WriteString(err.Key)
			return
		}
	} else {

		//c.StopRun()
		controllerName, _ := c.getControllerAndAction()

		redirectUri := c.Ctx.Input.Site() + ":" + library.Strval(c.Ctx.Input.Port())
		redirectUri += "/" + c.ModuleName
		redirectUri += "/" + controllerName
		redirectUri += "/callback"
		redirectUri += "?appid=" + appid
		redirectUri += "&scope=" + scope
		redirectUri += "&redirect=" + library.Urlencode(redirect)
		//c.Ctx.WriteString(redirectUri)
		//return

		// 根据Appid获取微信配置信息

		//配置微信参数
		config := &wechat.Config{
			AppID:          "your app id",
			AppSecret:      "your app secret",
			Token:          "your token",
			EncodingAESKey: "your encoding aes key",
		}
		wc := wechat.NewWechat(config)

		oauth := wc.GetOauth()
		err1 := oauth.Redirect(c.Ctx.ResponseWriter, c.Ctx.Request, redirectUri, scope, state)
		if err1 != nil {
			c.Ctx.WriteString(err1.Error())
			//fmt.Println(err)
			return
		}
	}

}

// 授权回调处理
func (c *SnsController) Callback() {

	appid := library.Trim(c.GetString("appid", ""))
	scope := library.Trim(c.GetString("scope", ""))
	redirect := library.Trim(c.GetString("redirect", ""))
	state := library.Trim(c.GetString("state", ""))
	code := library.Trim(c.GetString("code", ""))

	valid := validation.Validation{}
	valid.Required(appid, "appid Can not be empty.")
	valid.Required(scope, "scope Can not be empty.")
	valid.Required(redirect, "redirect Can not be empty.")
	valid.Required(state, "state Can not be empty.")
	valid.Required(code, "code Can not be empty.")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			c.Ctx.WriteString(err.Key)
			return
		}
	} else {
		//c.StopRun()

		// 根据Appid获取微信配置信息
		if code != "code" {
			//配置微信参数
			config := &wechat.Config{
				AppID:          "your app id",
				AppSecret:      "your app secret",
				Token:          "your token",
				EncodingAESKey: "your encoding aes key",
			}
			wc := wechat.NewWechat(config)
			oauth := wc.GetOauth()

			t1 := time.Now()
			timestamp := library.Strval(t1.Unix())

			resToken, err := oauth.GetUserAccessToken(code)
			if err != nil {
				//fmt.Println(err)
				c.Ctx.WriteString(err.Error())
				return
			}
			t1elapsed := time.Since(t1)

			if strings.Contains(redirect, "?") {
				redirect += "&FromUserName=" + resToken.OpenID
			} else {
				redirect += "?FromUserName=" + resToken.OpenID
			}
			redirect += "&t1elapsed=" + library.Strval(t1elapsed.Nanoseconds())
			redirect += "&timestamp=" + timestamp
			redirect += "&userToken=" + resToken.AccessToken
			redirect += "&refreshToken=" + resToken.RefreshToken
			redirect += "&signkey=" + c.getSignKey(resToken.OpenID, timestamp)

			if resToken.Scope == "snsapi_userinfo" || resToken.Scope == "snsapi_login" {
				t2 := time.Now()
				//getUserInfo
				userInfo, err := oauth.GetUserInfo(resToken.AccessToken, resToken.OpenID)
				if err != nil {
					//fmt.Println(err)
					c.Ctx.WriteString(err.Error())
					return
				}
				t2elapsed := time.Since(t2)

				redirect += "&nickname=" + library.Urlencode(userInfo.Nickname)
				redirect += "&headimgurl=" + library.Urlencode(userInfo.HeadImgURL)
				redirect += "&unionid=" + library.Urlencode(userInfo.Unionid)
				redirect += "&t2elapsed=" + library.Strval(t2elapsed.Nanoseconds())
				redirect += "&signkey=" + c.getSignKey(userInfo.Unionid, timestamp)
			}
		} else {
			t1 := time.Now()
			timestamp := library.Strval(t1.Unix())
			time.Sleep(time.Duration(2) * time.Second)
			t1elapsed := time.Since(t1)

			if strings.Contains(redirect, "?") {
				redirect += "&FromUserName=" + "guoyongrong"
			} else {
				redirect += "?FromUserName=" + "guoyongrong"
			}
			redirect += "&t1elapsed=" + library.Strval(t1elapsed.Nanoseconds())
			redirect += "&timestamp=" + timestamp
			redirect += "&userToken=" + "AccessToken"
			redirect += "&refreshToken=" + "RefreshToken"
			redirect += "&signkey=" + c.getSignKey("guoyongrong", timestamp)
		}

		c.Redirect(redirect, 302)
		return
	}

}
func (c *SnsController) getSignKey(p1 string, p2 string) string {
	return library.Sha1(p1 + "|" + "xxxx" + "|" + p2)
}
