package load_balance

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/credentials/insecure"

	userpb "github.com/liupch66/basic-go/grpc"
	"github.com/liupch66/basic-go/webook/pkg/netx"
)

type LoadBalancerTestSuite struct {
	suite.Suite
	etcdCli *clientv3.Client
}

func TestLoadBalancer(t *testing.T) {
	suite.Run(t, &LoadBalancerTestSuite{})
}

func (s *LoadBalancerTestSuite) SetupSuite() {
	client, err := clientv3.NewFromURL("127.0.0.1:22379")
	if err != nil {
		panic(err)
	}
	s.etcdCli = client
}

func (s *LoadBalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

func (s *LoadBalancerTestSuite) startServer(port string) {
	ctx := context.Background()
	em, err := endpoints.NewManager(s.etcdCli, "service/user")
	require.NoError(s.T(), err)
	// addr := netx.GetOutBoundIp() + ":8090"
	addr := netx.GetOutBoundIp() + port
	key := "service/user/" + addr
	leaseResp, err := s.etcdCli.Grant(ctx, 30)
	require.NoError(s.T(), err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]any{
			"cpu":    90,
			"weight": 100,
		},
	}, clientv3.WithLease(leaseResp.ID))

	kaCtx, kaCancel := context.WithCancel(context.Background())
	kaRespCh, err := s.etcdCli.KeepAlive(kaCtx, leaseResp.ID)
	require.NoError(s.T(), err)
	go func() {
		for kaResp := range kaRespCh {
			s.T().Log(kaResp.String(), time.Now().Second())
		}
	}()

	// lis, err := net.Listen("tcp", ":8090")
	lis, err := net.Listen("tcp", port)
	require.NoError(s.T(), err)
	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, &userpb.Server{Name: port})
	err = grpcServer.Serve(lis)
	require.NoError(s.T(), err)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	s.T().Log("Got signal: ", sig)

	// 这里到底怎么取消续约，查资料好像是 revoke
	// err = em.DeleteEndpoint(context.Background(), key, clientv3.WithLease(leaseResp.ID))
	// err = em.DeleteEndpoint(context.Background(), key)
	_, _ = s.etcdCli.Revoke(kaCtx, leaseResp.ID)
	kaCancel()

	require.NoError(s.T(), err)
	err = s.etcdCli.Close()
	require.NoError(s.T(), err)
	grpcServer.GracefulStop()
}

func (s *LoadBalancerTestSuite) TestClientPickFirst() {
	etcdResolver, err := resolver.NewBuilder(s.etcdCli)
	require.NoError(s.T(), err)
	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(etcdResolver),
		// gRPC 不指定，默认负载均衡算法是 pick first
		// 相当于 grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "pick_first"}`),

		// 指定轮询的两个写法，都是 json 字符串
		// 这些都在 google.golang.org/grpc@v1.69.2/balancer 下面各种负载均衡算法
		// 的 balancer.go 或类似文件 中找对应的 Name 常量
		// grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		// grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ]}`),

		// 指定 WRR，记得匿名引入这个包，触发里面的 init 方法（也就是要注册负载均衡算法）
		// 类似的还有 github.com/go-sql-driver/mysql 下面的 driver.go 里面的 init 方法
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "weighted_round_robin"}`),
	)
	require.NoError(s.T(), err)
	client := userpb.NewUserServiceClient(cc)

	for i := 0; i < 10; i++ {
		resp, err := client.GetById(context.Background(), &userpb.GetByIdReq{Id: 123})
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}
