package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/pkg/ginx/middleware/ratelimit"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisCli redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// cors 跨域资源共享
		cors.New(cors.Config{
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			ExposeHeaders:    []string{"x-jwt-token"},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		// 限流
		ratelimit.NewBuilder(redisCli, time.Minute, 100).Build(),
		// jwt 验证登录态
		middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/hello", "/users/signup",
			"/users/login", "/users/login_sms/code/send", "/users/login_sms").Build(),
	}
}
