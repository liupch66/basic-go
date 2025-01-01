package userpb

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestKratosTestSuite(t *testing.T) {
	suite.Run(t, new(KratosTestSuite))
}

type KratosTestSuite struct {
	suite.Suite
	etcdClient *clientv3.Client
}

func (s *KratosTestSuite) SetupSuite() {
	client, err := clientv3.NewFromURL("localhost:22379")
	require.NoError(s.T(), err)
	s.etcdClient = client
}

func (s *KratosTestSuite) TestServer() {
	grpcServer := grpc.NewServer(grpc.Address(":8090"), grpc.Middleware(recovery.Recovery()))
	RegisterUserServiceServer(grpcServer, &Server{})
	registry := etcd.New(s.etcdClient)
	app := kratos.New(kratos.Name("user"), kratos.Server(grpcServer), kratos.Registrar(registry))
	err := app.Run()
	require.NoError(s.T(), err)
}

func (s *KratosTestSuite) TestClient() {
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(), grpc.WithEndpoint("discovery:///user"), grpc.WithDiscovery(r))
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}
