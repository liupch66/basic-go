package grpcx

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"

	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/netx"
)

type Server struct {
	*grpc.Server
	Port int
	// 这些要导出，初始化时候使用
	Name     string
	EtcdAddr string
	L        logger.LoggerV1
	// EtcdAddrs []string
	etcdClient *clientv3.Client
	etcdKey    string
	epManager  endpoints.Manager
	kaCancel   func()
}

// NewServer 这里依赖注入的话，多个 service 共用一个 client 的时候，close 的时候会把其他 service 的 client 一起关了
// func NewServer(client *clientv3.Client){}

// Serve 启动服务器并且阻塞
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.kaCancel = cancel
	err = s.register(ctx)
	if err != nil {
		return err
	}
	return s.Server.Serve(lis)
}

func (s *Server) register(ctx context.Context) error {
	client, err := clientv3.NewFromURL(s.EtcdAddr)
	// client, err := clientv3.New(clientv3.Config{Endpoints: s.EtcdAddrs})
	if err != nil {
		return err
	}
	s.etcdClient = client

	em, err := endpoints.NewManager(client, "service/"+s.Name)
	if err != nil {
		return err
	}
	s.epManager = em

	// 创建租约
	var ttl int64 = 30
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	// 开启续约
	kaRespCh, err := client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range kaRespCh {
			s.L.Debug("续约", logger.String("resp", kaResp.String()))
		}
	}()

	addr := netx.GetOutBoundIp() + ":" + strconv.Itoa(s.Port)
	// e.g. service/user/192.168.135.108:8090
	s.etcdKey = fmt.Sprintf("service/%s/%s", s.Name, addr)
	return em.AddEndpoint(ctx, s.etcdKey, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.epManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.epManager.DeleteEndpoint(ctx, s.etcdKey)
		if err != nil {
			return err
		}
	}
	if s.etcdClient != nil {
		err := s.etcdClient.Close()
		if err != nil {
			return err
		}
	}
	s.Server.GracefulStop()
	return nil
}
