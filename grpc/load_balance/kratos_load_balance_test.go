package load_balance

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/random"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/liupch66/basic-go/grpc"
)

type KratosLBTestSuite struct {
	suite.Suite
	etcdClient *clientv3.Client
}

func (s *KratosLBTestSuite) SetupSuite() {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:22379"}})
	require.NoError(s.T(), err)
	s.etcdClient = cli
}

func (s *KratosLBTestSuite) TestClient() {
	// 默认是 WRR 负载均衡算法
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r))
	require.NoError(s.T(), err)
	defer cc.Close()

	client := userpb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &userpb.GetByIdReq{Id: 123})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *KratosLBTestSuite) TestClientLoadBalancer() {
	// 指定算法
	selector.SetGlobalSelector(random.NewBuilder())
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
		// 可以在这里传入节点的筛选器
		grpc.WithNodeFilter(func(ctx context.Context, nodes []selector.Node) []selector.Node {
			res := make([]selector.Node, 0, len(nodes))
			for _, node := range nodes {
				if node.Metadata()["vip"] == "true" {
					res = append(res, node)
				}
				// if node.Metadata()["vip"] == ctx.Value("is_vip") {
				//
				// }
			}
			if len(res) == 0 {
				return nodes
			}
			return res
		}),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := userpb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		// 通过查询用户，来设置标记位
		// ctx = context.WithValue(ctx, "is_vip", "true")
		resp, err := client.GetById(ctx, &userpb.GetByIdReq{Id: 123})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

// TestServer 启动服务器
func (s *KratosLBTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", "true")
	}()
	s.startServer(":8091", "false")
}

func (s *KratosLBTestSuite) startServer(addr string, vip string) {
	grpcSrv := grpc.NewServer(grpc.Address(addr), grpc.Middleware(recovery.Recovery()))
	userpb.RegisterUserServiceServer(grpcSrv, &userpb.Server{Name: addr})
	// etcd 注册中心
	r := etcd.New(s.etcdClient)
	app := kratos.New(kratos.Name("user"),
		kratos.Server(grpcSrv),
		kratos.Registrar(r),
		kratos.Metadata(map[string]string{"vip": vip, "region": "shanghai"}),
	)
	err := app.Run()
	require.NoError(s.T(), err)
}

func TestKratosLB(t *testing.T) {
	suite.Run(t, new(KratosLBTestSuite))
}
