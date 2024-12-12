package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"basic-go/webook/internal/web"
	ijwt "basic-go/webook/internal/web/jwt"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/pkg/ginx"
	"basic-go/webook/pkg/ginx/middleware/metrics"
	"basic-go/webook/pkg/ginx/middleware/ratelimit"
	"basic-go/webook/pkg/logger"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler,
	oauth2WechatHal *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHal.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	web.NewObservabilityHandler().RegisterRoutes(server)
	return server
}

func InitMiddlewares(l logger.LoggerV1, redisCli redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	// acBuilder := accesslog.NewMiddlewareBuilder(func(ctx context.Context, al *accesslog.AccessLog) {
	// 	l.Debug("HTTP", logger.Any("access_log", al))
	// })
	// viper.WatchConfig()
	// viper.OnConfigChange(func(in fsnotify.Event) {
	// 	// 假设 AccessLog 的 ReqBody 和 ReqBody 的动态开关 key 是 al_req_log 和 al_resp_log
	// 	acBuilder.AllowReqBody(viper.GetBool("al_req_log"))
	// 	acBuilder.AllowRespBody(viper.GetBool("al_resp_log"))
	// })
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geektime",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		// acBuilder.Build(),
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
			"/test/metrics",
		).Build(),
		(&metrics.PrometheusBuilder{
			Namespace:  "geektime",
			Subsystem:  "webook",
			Name:       "gin_http",
			Help:       "统计 GIN HTTP 接口",
			InstanceID: "my-instance-1",
		}).Build(),
		otelgin.Middleware("webook"),
	}
}
