package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	igrpc "github.com/liupch66/basic-go/webook/interact/grpc"
	"github.com/liupch66/basic-go/webook/pkg/grpcx"
)

func InitGRPCxServer(interactSrv *igrpc.InteractServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("grpc.server", &cfg); err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	interactSrv.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
