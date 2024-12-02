//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	ijwt "basic-go/webook/internal/web/jwt"
	"basic-go/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		dao.NewUserDAO,
		cache.NewUserCache, cache.NewCodeCache,
		repository.NewUserRepository, repository.NewCodeRepository,
		service.NewUserService, service.NewCodeService, ioc.InitSmsService, ioc.InitWechatService,
		web.NewUserHandler, ioc.InitWechatHandlerConfig, web.NewOAuth2WechatHandler, ijwt.NewRedisJwtHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}
