package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/web"
	ijwt "basic-go/webook/internal/web/jwt"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/pkg/ginx/middleware/ratelimit"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler, oauth2WechatHal *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHal.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisCli redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// cors 跨域资源共享
		cors.New(cors.Config{
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
					return true
				}
				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		// 限流
		ratelimit.NewBuilder(redisCli, time.Minute, 100).Build(),
		// jwt 验证登录态
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).IgnorePaths(
			"/hello",
			"/users/signup",
			"/users/login",
			"/users/login_sms/code/send",
			"/users/login_sms",
			"/oauth2/wechat/authurl",
			"/oauth2/wechat/callback",
			"/wechat/callback.do",
			// access_token 过期了要通过 refresh_token 刷新
			"/users/refresh_token",
		).Build(),
	}
}
