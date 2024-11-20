package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"basic-go/webook/config"
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUserHandler(db)
	// 注册用户相关接口路由
	u.RegisterRoutes(server)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, world!")
	})
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initDB() *gorm.DB {
	// GORM 连接数据库
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	// 建表
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	// 解决跨域问题 CORS
	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:3000"},
		// AllowMethods:     []string{"POST"},
		// 指定客户端在跨域请求中允许发送的自定义请求头，告诉浏览器，哪些请求头是允许随请求发送到服务器的
		AllowHeaders: []string{"Authorization", "Content-Type"},
		// 指定客户端可以在响应中访问的自定义响应头，告诉浏览器，在跨域响应中，哪些头部可以被 JavaScript 代码访问
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// cookie: 基于 cookie 的实现
	// store := cookie.NewStore([]byte("secret"))
	// memstore: 基于内存的实现，单机单实例部署
	// store := memstore.NewStore([]byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"), []byte("1x6`djgK$0KM].Sz:SqLa?BF=OJhuIRG"))
	// redis: 基于 redis 的实现，多实例部署
	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	// 	[]byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"), []byte("1x6`djgK$0KM].Sz:SqLa?BF=OJhuIRG"))
	// if err != nil {
	// 	panic(err)
	// }
	// 设置 cookie 的键值对，ssid: sessionID（由服务器自动生成，是一个加密的标识符）
	// server.Use(sessions.Sessions("ssid", store))
	// session 机制 登录校验
	// server.Use(middleware.NewLoginMiddlewareBuilder().Build())
	// jwt 机制 登录校验
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/hello", "/users/signup", "/users/login").Build())
	// redisCli := redis.NewClient(&redis.Options{
	// 	Addr:     config.Config.Redis.Addr,
	// 	Password: config.Config.Redis.Password,
	// 	// Redis 提供多个逻辑数据库（默认 16 个，编号从 0 到 15）。每个数据库是独立的，但它们共享同一个实例的资源（如内存）。
	// 	DB: config.Config.Redis.DB,
	// })
	// server.Use(ratelimit.NewBuilder(redisCli, time.Minute, 100).Build())
	return server
}

func initUserHandler(db *gorm.DB) *web.UserHandler {
	// 初始化 UserHandler
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}
