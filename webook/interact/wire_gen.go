// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitApp() *app {
	db := ioc.InitDB()
	interactDAO := dao.NewGORMInteractDAO(db)
	cmdable := ioc.InitRedis()
	interactCache := cache.NewRedisInteractCache(cmdable)
	loggerV1 := ioc.InitLogger()
	interactRepository := repository.NewCachedInteractRepository(interactDAO, interactCache, loggerV1)
	interactService := service.NewInteractService(interactRepository, loggerV1)
	interactServiceServer := grpc.NewInteractServiceServer(interactService)
	server := ioc.InitGrpcxServer(interactServiceServer)
	client := ioc.InitKafka()
	interactReadEventConsumer := events.NewInteractReadEventConsumer(client, interactRepository, loggerV1)
	v := ioc.NewConsumers(interactReadEventConsumer)
	mainApp := &app{
		server:    server,
		consumers: v,
	}
	return mainApp
}

// wire.go:

var thirdPartyProvider = wire.NewSet(ioc.InitDB, ioc.InitRedis, ioc.InitLogger, ioc.InitKafka)

var interactServiceProvider = wire.NewSet(dao.NewGORMInteractDAO, cache.NewRedisInteractCache, repository.NewCachedInteractRepository, service.NewInteractService)
