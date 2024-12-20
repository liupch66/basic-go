//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/liupch66/basic-go/webook/interact/events"
	"github.com/liupch66/basic-go/webook/interact/grpc"
	"github.com/liupch66/basic-go/webook/interact/ioc"
	"github.com/liupch66/basic-go/webook/interact/repository"
	"github.com/liupch66/basic-go/webook/interact/repository/cache"
	"github.com/liupch66/basic-go/webook/interact/repository/dao"
	"github.com/liupch66/basic-go/webook/interact/service"
)

var thirdPartyProvider = wire.NewSet(
	ioc.InitDB, ioc.InitRedis, ioc.InitLogger, ioc.InitKafka,
)

var interactServiceProvider = wire.NewSet(
	dao.NewGORMInteractDAO, cache.NewRedisInteractCache,
	repository.NewCachedInteractRepository,
	service.NewInteractService,
)

func InitApp() *app {
	wire.Build(
		thirdPartyProvider, interactServiceProvider,
		events.NewInteractReadEventConsumer, ioc.NewConsumers,
		grpc.NewInteractServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(app), "*"),
	)
	return new(app)
}
