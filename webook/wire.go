//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/article"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	articleDAO "basic-go/webook/internal/repository/dao/article"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	ijwt "basic-go/webook/internal/web/jwt"
	"basic-go/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		dao.NewUserDAO,
		cache.NewUserCache, cache.NewCodeCache, articleDAO.NewGORMArticleDAO,
		repository.NewUserRepository, repository.NewCodeRepository, article.NewCachedArticleRepository,
		service.NewUserService, service.NewCodeService, ioc.InitSmsService, ioc.InitWechatService, service.NewArticleService,
		web.NewUserHandler, ioc.InitWechatHandlerConfig, web.NewOAuth2WechatHandler, ijwt.NewRedisJwtHandler, web.NewArticleHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}
