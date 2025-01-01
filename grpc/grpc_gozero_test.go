package userpb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}

type GoZeroTestSuite struct {
	suite.Suite
}

func (s *GoZeroTestSuite) TestGoZeroServer() {
	// 正常来说，这个都是从配置文件中读取的，类似 main 函数那样，从命令行接受配置文件的路径
	// configFile := flag.String("f", "etc/consul.yaml", "the config file")
	// var c zrpc.RpcServerConf
	// conf.MustLoad(*configFile, &c)
	c := zrpc.RpcServerConf{
		ListenOn: ":8090",
		Etcd: discov.EtcdConf{
			// Etcd 集群的地址
			Hosts: []string{"localhost:22379"},
			// 服务注册的 key，用于在 Etcd 中标识服务
			Key: "user",
		},
	}
	server, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		RegisterUserServiceServer(grpcServer, &Server{})
	})
	require.NoError(s.T(), err)
	server.Start()
}

func (s *GoZeroTestSuite) TestGoZeroClient() {
	c := zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:22379"},
			Key:   "user",
		},
	}
	// zClient, err := zrpc.NewClient(c)
	zClient := zrpc.MustNewClient(c)
	client := NewUserServiceClient(zClient.Conn())
	resp, err := client.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}
