//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"github.com/liupch66/basic-go/webook/interact/grpc"
	"github.com/liupch66/basic-go/webook/interact/repository"
	"github.com/liupch66/basic-go/webook/interact/repository/cache"
	"github.com/liupch66/basic-go/webook/interact/repository/dao"
	"github.com/liupch66/basic-go/webook/interact/service"
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

func InitGrpcServer() *grpc.InteractServiceServer {
	wire.Build(thirdPS, interactSvcPS, grpc.NewInteractServiceServer)
	return new(grpc.InteractServiceServer)
}
