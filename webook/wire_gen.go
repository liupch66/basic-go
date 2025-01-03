// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	article3 "github.com/liupch66/basic-go/webook/internal/events/article"
	"github.com/liupch66/basic-go/webook/internal/repository"
	article2 "github.com/liupch66/basic-go/webook/internal/repository/article"
	"github.com/liupch66/basic-go/webook/internal/repository/cache"
	"github.com/liupch66/basic-go/webook/internal/repository/dao"
	"github.com/liupch66/basic-go/webook/internal/repository/dao/article"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/internal/web"
	"github.com/liupch66/basic-go/webook/internal/web/jwt"
	"github.com/liupch66/basic-go/webook/ioc"
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
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewCachedArticleRepository(userRepository, articleDAO, articleCache, loggerV1)
	client := ioc.InitKafka()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article3.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	clientv3Client := ioc.InitEtcdClient()
	interactServiceClient := ioc.InitInteractGRPCClientV1(clientv3Client)
	articleHandler := web.NewArticleHandler(articleService, interactServiceClient, loggerV1)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	v2 := ioc.NewConsumers()
	rankLocalCache := cache.NewRankLocalCache()
	redisRankCache := cache.NewRedisRankCache(cmdable)
	rankRepository := repository.NewCachedRankRepository(rankLocalCache, redisRankCache)
	rankService := service.NewBatchRankService(articleService, interactServiceClient, rankRepository)
	rlockClient := ioc.InitRLockClient(cmdable)
	rankJob := ioc.InitRankJob(rankService, rlockClient, loggerV1)
	cron := ioc.InitJobs(loggerV1, rankJob)
	app := &App{
		web:       engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

var rankServiceSet = wire.NewSet(cache.NewRedisRankCache, cache.NewRankLocalCache, repository.NewCachedRankRepository, service.NewBatchRankService)
