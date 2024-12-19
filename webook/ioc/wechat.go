package ioc

import (
	"basic-go/webook/internal/service/oauth2/wechat"
	"basic-go/webook/internal/web"
	"basic-go/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	// 一号店
	appId := "wxbdc5610cc59c1631"
	// appId, ok := os.LookupEnv("WECHAT_APP_ID")
	// if !ok {
	// 	panic("没有找到环境变量 WECHAT_APP_ID")
	// }
	// appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	// if !ok {
	// 	panic("没有找到环境变量 WECHAT_APP_SECRET")
	// }
	return wechat.NewWechatService(appId, "", l)
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{Secure: false}
}
