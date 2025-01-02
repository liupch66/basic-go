package weightedroundrobin

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "github.com/liupch66/basic-go/grpc"
	"github.com/liupch66/basic-go/webook/pkg/netx"
)

type BalancerTestSuite struct {
	suite.Suite
	etcdCli *clientv3.Client
}

func TestBalancer(t *testing.T) {
	suite.Run(t, &BalancerTestSuite{})
}

func (s *BalancerTestSuite) SetupSuite() {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:22379"},
	})
	require.NoError(s.T(), err)
	s.etcdCli = etcdCli
}

// 这里步骤都比较粗糙，没有什么租约，续约，优雅退出什么的，webook/grpc/load_balance/load_balancer_test.go 比较详细
func (s *BalancerTestSuite) startServer(port string, weight int) {
	em, err := endpoints.NewManager(s.etcdCli, "service/user")
	require.NoError(s.T(), err)
	addr := netx.GetOutBoundIp() + port
	key := "service/user/" + addr
	err = em.AddEndpoint(context.Background(), key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
			// "cpu": 90,
		},
	})
	require.NoError(s.T(), err)

	grpcSrv := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcSrv, &userpb.Server{Name: port})
	lis, err := net.Listen("tcp", port)
	require.NoError(s.T(), err)
	err = grpcSrv.Serve(lis)
	require.NoError(s.T(), err)
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 10)
	}()
	s.startServer(":8091", 20)
}

func (s *BalancerTestSuite) TestClient() {
	etcdResolver, err := resolver.NewBuilder(s.etcdCli)
	require.NoError(s.T(), err)

	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(etcdResolver),
		// 其他包要用这个自定义算法要匿名引入 webook/pkg/grpcx/balancer/weightedroundrobin，执行 init 方法
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "custom_weighted_round_robin"}`),
		// grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "custom_weighted_round_robin": {} } ]}`),
	)
	require.NoError(s.T(), err)
	client := userpb.NewUserServiceClient(cc)

	for i := 0; i < 12; i++ {
		resp, err := client.GetById(context.Background(), &userpb.GetByIdReq{Id: 1234})
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}
