//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	article3 "basic-go/webook/internal/events/article"
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/article"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	article2 "basic-go/webook/internal/repository/dao/article"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/service/oauth2/wechat"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/jwt"
	"basic-go/webook/ioc"
)

var (
	thirdPS = wire.NewSet(InitTestDB, InitRedis, InitLog,
		InitKafka, ioc.InitSyncProducer, article3.NewSaramaSyncProducer)
	userSvcPS = wire.NewSet(dao.NewUserDAO, cache.NewUserCache,
		repository.NewUserRepository,
		service.NewUserService)
	codeSvcPS = wire.NewSet(cache.NewCodeCache,
		repository.NewCodeRepository, ioc.InitSmsService,
		service.NewCodeService)
	articleSvcPS = wire.NewSet(article2.NewGORMArticleDAO, cache.NewRedisArticleCache,
		article.NewCachedArticleRepository,
		service.NewArticleService)
	wechatSvcPS   = wire.NewSet(ioc.InitWechatService)
	interactSvcPS = wire.NewSet(dao.NewGORMInteractDAO, cache.NewRedisInteractCache,
		repository.NewCachedInteractRepository,
		service.NewInteractService)
)

func InitUserSvc() service.UserService {
	wire.Build(thirdPS, userSvcPS)
	return service.NewUserService(nil, nil)
}

func InitCodeSvc() service.CodeService {
	wire.Build(thirdPS, codeSvcPS)
	return service.NewCodeService(nil, nil)
}

func InitWechatSvc() wechat.Service {
	wire.Build(thirdPS, wechatSvcPS)
	return wechat.NewWechatService("", "", nil)
}

func InitInteractService() service.InteractService {
	wire.Build(thirdPS, interactSvcPS)
	return service.NewInteractService(nil, nil)
}

// 这里注入 artDAO 是为了方便集成测试 GORM DB 和 MongoDB 实现的文章储存
func InitArticleHandler(artDAO article2.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdPS, interactSvcPS, userSvcPS,
		cache.NewRedisArticleCache,
		article.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPS, userSvcPS, codeSvcPS, wechatSvcPS,
		article2.NewGORMArticleDAO,
		// article2.NewMongoDBDAO, InitMongoDB, // 方便切换成 mongoDB
		jwt.NewRedisJwtHandler, ioc.InitWechatHandlerConfig,
		ioc.InitMiddlewares,
		web.NewUserHandler, web.NewOAuth2WechatHandler, InitArticleHandler, // 这里注入 InitArticleHandler 是为了方便测试
		ioc.InitWebServer,
	)
	return &gin.Engine{}
}
