//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/liupch66/basic-go/webook/interact/events"
	repository2 "github.com/liupch66/basic-go/webook/interact/repository"
	cache2 "github.com/liupch66/basic-go/webook/interact/repository/cache"
	dao2 "github.com/liupch66/basic-go/webook/interact/repository/dao"
	service2 "github.com/liupch66/basic-go/webook/interact/service"
	article2 "github.com/liupch66/basic-go/webook/internal/events/article"
	"github.com/liupch66/basic-go/webook/internal/repository"
	"github.com/liupch66/basic-go/webook/internal/repository/article"
	"github.com/liupch66/basic-go/webook/internal/repository/cache"
	"github.com/liupch66/basic-go/webook/internal/repository/dao"
	article3 "github.com/liupch66/basic-go/webook/internal/repository/dao/article"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/internal/web"
	ijwt "github.com/liupch66/basic-go/webook/internal/web/jwt"
	"github.com/liupch66/basic-go/webook/ioc"
)

var rankServiceSet = wire.NewSet(
	cache.NewRedisRankCache,
	cache.NewRankLocalCache,
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
		service.NewArticleService, service2.NewInteractService, ioc.InitInteractGRPCClient,

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
