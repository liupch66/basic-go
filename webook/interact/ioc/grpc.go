package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	igrpc "github.com/liupch66/basic-go/webook/interact/grpc"
	"github.com/liupch66/basic-go/webook/pkg/grpcx"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

func InitGRPCxServer(interactSrv *igrpc.InteractServiceServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("grpc.server", &cfg); err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	interactSrv.Register(server)

	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		EtcdAddr: cfg.EtcdAddr,
		Name:     "interact",
		L:        l,
	}
}
