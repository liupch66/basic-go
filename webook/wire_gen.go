// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	article3 "basic-go/webook/internal/events/article"
	"basic-go/webook/internal/repository"
	article2 "basic-go/webook/internal/repository/article"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/repository/dao/article"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/jwt"
	"basic-go/webook/ioc"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitApp() *App {
	loggerV1 := ioc.InitLogger()
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJwtHandler(cmdable)
	v := ioc.InitMiddlewares(loggerV1, cmdable, handler)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitWechatService(loggerV1)
	wechatHandlerConfig := ioc.InitWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, wechatHandlerConfig, handler)
	articleDAO := article.NewGORMArticleDAO(db)
	articleRepository := article2.NewCachedArticleRepository(articleDAO, loggerV1)
	client := ioc.InitKafka()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article3.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	interactDAO := dao.NewGORMInteractDAO(db)
	interactCache := cache.NewRedisInteractCache(cmdable)
	interactRepository := repository.NewCachedInteractRepository(interactDAO, interactCache, loggerV1)
	interactService := service.NewInteractService(interactRepository, loggerV1)
	articleHandler := web.NewArticleHandler(articleService, interactService, loggerV1)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactReadEventBatchConsumer := article3.NewInteractReadEventBatchConsumer(client, interactRepository, loggerV1)
	v2 := ioc.NewConsumers(interactReadEventBatchConsumer)
	app := &App{
		web:       engine,
		consumers: v2,
	}
	return app
}
