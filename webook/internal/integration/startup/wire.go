//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

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

var thirdPS = wire.NewSet(InitTestDB, InitRedis, InitLog)

var userSvcPS = wire.NewSet(
	dao.NewUserDAO, cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService,
)

var codeSvcPS = wire.NewSet(
	cache.NewCodeCache,
	repository.NewCodeRepository, ioc.InitSmsService,
	service.NewCodeService,
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
	wire.Build(thirdPS, ioc.InitWechatService)
	return wechat.NewService("", "", nil)
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(thirdPS, article2.NewGORMArticleDAO, article.NewCachedArticleRepository,
		service.NewArticleService, web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPS,
		InitUserSvc, InitCodeSvc, InitWechatSvc,
		jwt.NewRedisJwtHandler, ioc.InitWechatHandlerConfig,
		ioc.InitMiddlewares,
		web.NewUserHandler, web.NewOAuth2WechatHandler, InitArticleHandler,
		ioc.InitWebServer,
	)
	return &gin.Engine{}
}
