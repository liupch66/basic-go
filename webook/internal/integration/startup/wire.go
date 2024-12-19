//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	repository2 "github.com/liupch66/basic-go/webook/interact/repository"
	cache2 "github.com/liupch66/basic-go/webook/interact/repository/cache"
	dao2 "github.com/liupch66/basic-go/webook/interact/repository/dao"
	service2 "github.com/liupch66/basic-go/webook/interact/service"
	article3 "github.com/liupch66/basic-go/webook/internal/events/article"
	"github.com/liupch66/basic-go/webook/internal/repository"
	"github.com/liupch66/basic-go/webook/internal/repository/article"
	"github.com/liupch66/basic-go/webook/internal/repository/cache"
	"github.com/liupch66/basic-go/webook/internal/repository/dao"
	article2 "github.com/liupch66/basic-go/webook/internal/repository/dao/article"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/internal/service/oauth2/wechat"
	"github.com/liupch66/basic-go/webook/internal/web"
	"github.com/liupch66/basic-go/webook/internal/web/jwt"
	"github.com/liupch66/basic-go/webook/ioc"
)

var (
	thirdPS = wire.NewSet(InitTestDB, InitRedis, InitLog,
		InitKafka, InitSyncProducer, article3.NewSaramaSyncProducer)
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
	interactSvcPS = wire.NewSet(dao2.NewGORMInteractDAO, cache2.NewRedisInteractCache,
		repository2.NewCachedInteractRepository,
		service2.NewInteractService)
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

func InitInteractService() service2.InteractService {
	wire.Build(thirdPS, interactSvcPS)
	return service2.NewInteractService(nil, nil)
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
