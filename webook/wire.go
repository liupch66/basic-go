//go:build wireinject

package main

import (
	"github.com/google/wire"

	"basic-go/webook/interact/events"
	repository2 "basic-go/webook/interact/repository"
	cache2 "basic-go/webook/interact/repository/cache"
	dao2 "basic-go/webook/interact/repository/dao"
	service2 "basic-go/webook/interact/service"
	article2 "basic-go/webook/internal/events/article"
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/article"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	article3 "basic-go/webook/internal/repository/dao/article"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	ijwt "basic-go/webook/internal/web/jwt"
	"basic-go/webook/ioc"
)

var rankServiceSet = wire.NewSet(
	cache.NewRedisRankCache,
	cache.NewLocalRankCache,
	repository.NewCachedRankRepository,
	service.NewBatchRankService,
)

func InitApp() *App {
	wire.Build(
		ioc.InitDB, ioc.InitRedis, ioc.InitRLockClient, ioc.InitLogger,
		ioc.InitKafka, ioc.InitSyncProducer, article2.NewSaramaSyncProducer,
		events.NewInteractReadEventBatchConsumer, ioc.NewConsumers,

		dao.NewUserDAO, article3.NewGORMArticleDAO, dao2.NewGORMInteractDAO,
		cache.NewUserCache, cache.NewCodeCache, cache2.NewRedisInteractCache, cache.NewRedisArticleCache,

		repository.NewUserRepository, repository.NewCodeRepository, article.NewCachedArticleRepository,
		repository2.NewCachedInteractRepository,

		service.NewUserService, service.NewCodeService, ioc.InitSmsService, ioc.InitWechatService,
		service.NewArticleService, service2.NewInteractService,

		rankServiceSet,
		ioc.InitRankJob,
		ioc.InitJobs,

		web.NewUserHandler, ioc.InitWechatHandlerConfig, web.NewOAuth2WechatHandler, ijwt.NewRedisJwtHandler,
		web.NewArticleHandler,

		ioc.InitMiddlewares,

		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
