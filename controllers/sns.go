package controllers

import (
	"go_weixin_module/library"
	"strings"
	//"errors"
	//"fmt"
	//"net"
	//"net/url"
	"time"

	"github.com/astaxie/beego"
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
	refresh := library.Intval(c.GetString("refresh", "0"))

	secret := scope + "_" + appid

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
		if refresh == 0 {
			openID, _ := c.GetSecureCookie(secret, secret+"_openid")
			if !library.Empty(openID) {
				t1 := time.Now()
				timestamp := library.Strval(t1.Unix())
				userToken, _ := c.GetSecureCookie(secret, secret+"_userToken")
				refreshToken, _ := c.GetSecureCookie(secret, secret+"_refreshToken")
				nickname, _ := c.GetSecureCookie(secret, secret+"_nickname")
				headimgurl, _ := c.GetSecureCookie(secret, secret+"_headimgurl")
				unionid, _ := c.GetSecureCookie(secret, secret+"_unionid")
				t1elapsed := time.Since(t1)

				others := make(map[string]string, 0)
				others["t1elapsed"] = library.Strval(t1elapsed.Nanoseconds())
				others["timestamp"] = timestamp
				others["userToken"] = library.Urlencode(userToken)
				others["refreshToken"] = library.Urlencode(refreshToken)
				others["signkey"] = c.getSignKey(openID, timestamp)

				others["nickname"] = library.Urlencode(nickname)
				others["headimgurl"] = library.Urlencode(headimgurl)
				others["unionid"] = library.Urlencode(unionid)
				others["signkey2"] = c.getSignKey(unionid, timestamp)
				others["t2elapsed"] = library.Strval(0)
				others["is_cookie_from"] = library.Strval(1)
				redirect = c.getRedirectUrl(redirect, openID, others)

				c.Redirect(redirect, 302)
				return
			}

		}

		//c.StopRun()
		controllerName, _ := c.getControllerAndAction()

		//redirectUri := c.Ctx.Input.Site() + ":" + library.Strval(c.Ctx.Input.Port())
		redirectUri := c.Ctx.Input.Site()
		redirectUri += "/" + c.ModuleName
		redirectUri += "/" + controllerName
		redirectUri += "/callback"
		redirectUri += "?appid=" + appid
		redirectUri += "&scope=" + scope
		//redirectUri += "&redirect=" + library.Urlencode(redirect)
		c.SetSession("redirect", redirect)

		//c.Ctx.WriteString(redirectUri)
		//return

		// 根据Appid获取微信配置信息

		//配置微信参数
		config := &wechat.Config{
			AppID:          beego.AppConfig.String("weixinappid"),
			AppSecret:      beego.AppConfig.String("weixinappsecret"),
			Token:          "",
			EncodingAESKey: "",
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
	//redirect := library.Trim(c.GetString("redirect", ""))
	redirect := library.Strval(c.GetSession("redirect"))

	state := library.Trim(c.GetString("state", ""))
	code := library.Trim(c.GetString("code", ""))
	secret := scope + "_" + appid

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
				AppID:          beego.AppConfig.String("weixinappid"),
				AppSecret:      beego.AppConfig.String("weixinappsecret"),
				Token:          "",
				EncodingAESKey: "",
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

			// 设置cookie
			c.SetSecureCookie(secret, secret+"_openid", resToken.OpenID, 3600*1.5, "/")
			c.SetSecureCookie(secret, secret+"_userToken", resToken.AccessToken, 3600*1.5, "/")
			c.SetSecureCookie(secret, secret+"_refreshToken", resToken.RefreshToken, 3600*1.5, "/")

			others := make(map[string]string, 0)
			others["t1elapsed"] = library.Strval(t1elapsed.Nanoseconds())
			others["timestamp"] = timestamp
			others["userToken"] = library.Urlencode(resToken.AccessToken)
			others["refreshToken"] = library.Urlencode(resToken.RefreshToken)
			others["signkey"] = c.getSignKey(resToken.OpenID, timestamp)

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

				// 设置cookie
				c.SetSecureCookie(secret, secret+"_nickname", userInfo.Nickname, 3600*1.5, "/")
				c.SetSecureCookie(secret, secret+"_headimgurl", userInfo.HeadImgURL, 3600*1.5, "/")
				c.SetSecureCookie(secret, secret+"_unionid", userInfo.Unionid, 3600*1.5, "/")

				others["nickname"] = library.Urlencode(userInfo.Nickname)
				others["headimgurl"] = library.Urlencode(userInfo.HeadImgURL)
				others["unionid"] = userInfo.Unionid
				others["t2elapsed"] = library.Strval(t2elapsed.Nanoseconds())
				others["signkey2"] = c.getSignKey(userInfo.Unionid, timestamp)
			}

			redirect = c.getRedirectUrl(redirect, resToken.OpenID, others)
		} else {
			t1 := time.Now()
			timestamp := library.Strval(t1.Unix())
			time.Sleep(time.Duration(2) * time.Second)
			t1elapsed := time.Since(t1)

			others := make(map[string]string, 0)
			others["t1elapsed"] = library.Strval(t1elapsed.Nanoseconds())
			others["timestamp"] = timestamp
			others["userToken"] = "AccessToken"
			others["refreshToken"] = "RefreshToken"
			others["signkey"] = c.getSignKey("guoyongrong", timestamp)
			redirect = c.getRedirectUrl(redirect, "guoyongrong", others)
		}

		c.Redirect(redirect, 302)
		return
	}

}
func (c *SnsController) getSignKey(p1 string, p2 string) string {
	return library.Sha1(p1 + "|" + beego.AppConfig.String("weixinsignkey") + "|" + p2)
}

func (c *SnsController) getRedirectUrl(redirect string, openid string, others map[string]string) string {
	if strings.Contains(redirect, "?") {
		redirect += "&FromUserName=" + openid
	} else {
		redirect += "?FromUserName=" + openid
	}
	for key, val := range others {
		redirect += "&" + key + "=" + val
	}
	return redirect
}
