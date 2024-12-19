//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"basic-go/webook/interact/repository"
	"basic-go/webook/interact/repository/cache"
	"basic-go/webook/interact/repository/dao"
	"basic-go/webook/interact/service"
)

var thirdPS = wire.NewSet(InitTestDB, InitRedis, InitLog)

var interactSvcPS = wire.NewSet(
	dao.NewGORMInteractDAO, cache.NewRedisInteractCache,
	repository.NewCachedInteractRepository,
	service.NewInteractService,
)

func InitInteractService() service.InteractService {
	wire.Build(thirdPS, interactSvcPS)
	return service.NewInteractService(nil, nil)
}
