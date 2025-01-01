package userpb

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/liupch66/basic-go/webook/pkg/netx"
)

func TestEtcd(t *testing.T) {
	suite.Run(t, &EtcdTestSuite{})
}

type EtcdTestSuite struct {
	suite.Suite
	client *clientv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	// client, err := clientv3.NewFromURL("localhost:22379")
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:22379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

// 关键点：
// • 创建一个 endpoint.Manager。
// • 添加一个 Endpoint。
// • 后续如果注册数据有变动，那么就可以调用 Update 方法。
// • 后续在退出的时候，需要调用一个 Delete 方法，删除这个 Endpoint
func (s *EtcdTestSuite) TestServer() {

	// endpoint 以服务为维度。一个服务一个 Manager。
	// Manager can be used to add/ remove & inspect endpoints stored in etcd for a particular target.
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)
	// ....在这一步之前完成所以的启动的准备工作，包括缓存预加载之类的事情

	// 创建续约
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 过了 1/3 ttl 就开始续约，源码：nextKeepAlive := time.Now().Add((time.Duration(karesp.TTL) * time.Second) / 3.0)
	var ttl int64 = 30 // 30s
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)

	// addr := "127.0.0.1:8090"
	addr := netx.GetOutBoundIp() + ":8090" // "192.168.135.108:8090"
	// key 是指这个 service 实例的 key，用 instance id，没有就 本机 IP + port
	key := "service/user/" + addr
	// AddEndpoint registers a single endpoint in etcd.
	// endpoints: endpoint key should be prefixed with '$target/'
	err = em.AddEndpoint(context.Background(), key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]any{
			"weight": 100,
		},
	}, clientv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)
	// 命令行 etcdctl --endpoints=http://127.0.0.1:22379 get service/user/127.0.0.1:8090 能查看到信息

	// 操作续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		kaRespCh, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)
		for resp := range kaRespCh {
			// 正常就是打印一下 DEBUG 日志啥的
			s.T().Log(resp.String())
			s.T().Log(time.Now().Second())
		}
	}()

	// 万一注册信息变动
	go func() {
		tk := time.NewTicker(time.Second)
		for {
			for now := range tk.C {
				// 方法一
				// upsert 或者 set 的语义
				err := em.AddEndpoint(context.Background(), key, endpoints.Endpoint{
					Addr: addr,
					Metadata: map[string]any{
						"weight": 200,
						"time":   now,
					},
				}, clientv3.WithLease(leaseResp.ID))
				if err != nil {
					s.T().Log(err)
				}

				// 方法二
				// metadata 一般用在客户端
				// err := em.Update(context.Background(), []*endpoints.UpdateWithOpts{
				// 	{
				// 		Update: endpoints.Update{
				// 			// 这里 Op 只有 Add 和 Delete
				// 			Op:  endpoints.Add,
				// 			Key: key,
				// 			Endpoint: endpoints.Endpoint{
				// 				Port:     addr,
				// 				Metadata: now.Format(time.DateTime),
				// 			},
				// 		},
				// 	},
				// })
				// if err != nil {
				// 	s.T().Log(err)
				// }
			}
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)
	err = server.Serve(lis)
	s.T().Log(err)

	// 服务退出
	kaCancel()
	err = em.DeleteEndpoint(context.Background(), key)
	assert.NoError(s.T(), err)
	s.client.Close()
	server.GracefulStop()
}

func (s *EtcdTestSuite) TestClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// 常见的 URL 格式一般是：协议://主机名:端口/路径，即 scheme://host:port/path，例如：http://example.com/path/to/resource
	// 这里相当于空的主机和端口, etcd://nil/service/user,
	// etcd 是 scheme， /service/user 是 path，指向在 Etcd 中注册的服务路径，表示目标服务所在的位置。
	// 从这个 path 去掉前缀 "/"，得到 endpoint = service/user
	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}
