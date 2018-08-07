// 依赖第3方库
go get -u github.com/silenceper/wechat
https://beego.me/docs/intro/
bee run

http://192.168.5.80:58081/weixin/index/index
http://192.168.5.80:58081/weixin/sns/index?appid=xxx&scope=snsapi_userinfo&state=test&redirect=https%3A%2F%2Fwww.baidu.com%2F
http://192.168.5.80:58081/weixin/sns/index?appid=xxx&scope=snsapi_userinfo&state=test&redirect=http%3A%2F%2Fwww.baidu.com%2F%3Fa%3D1
http://192.168.5.80:58081/weixin/sns/callback?appid=xxx&scope=snsapi_userinfo&state=test&redirect=http%3A%2F%2Fwww.baidu.com%2F%3Fa%3D1&code=code
