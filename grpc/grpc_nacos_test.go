package userpb

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/liupch66/basic-go/webook/pkg/netx"
)

func newNamingClient() naming_client.INamingClient {
	clientConfig := constant.ClientConfig{
		// 这是在 http://localhost:8848/nacos 创建的命名空间 id
		NamespaceId:         "77955bab-3215-48c4-b2cf-3e04db63426e",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		CacheDir:            "/tmp/nacos/cache",
		LogDir:              "/tmp/nacos/log",
		LogLevel:            "debug",
	}
	serverConfigs := []constant.ServerConfig{{IpAddr: "localhost", Port: 8848}}
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		panic(err)
	}
	return namingClient
}

func TestNacosServer(t *testing.T) {
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	RegisterUserServiceServer(grpcServer, &Server{})

	namingClient := newNamingClient()
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		// Ip:          "localhost",
		Ip:          netx.GetOutBoundIp(),
		Port:        8090,
		ServiceName: "user",
		GroupName:   "group-a",
		ClusterName: "cluster-a",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
	})
	t.Logf("success: %t, err: %v\n", success, err)
	if !success || err != nil {
		t.Fatal("register Service Instance failed!")
	}

	// 启动gRPC服务器监听
	err = grpcServer.Serve(lis)
	require.NoError(t, err)

	// 优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	t.Log("Shutting down...")
	_, err = namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          "localhost",
		Port:        8090,
		ServiceName: "user",
		GroupName:   "group-a",
		Cluster:     "cluster-a",
		Ephemeral:   true,
	})
	require.NoError(t, err)
	grpcServer.GracefulStop()
}

func TestNacosClient(t *testing.T) {
	namingClient := newNamingClient()
	instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		Clusters:    []string{"cluster-a"},
		GroupName:   "group-a",
		ServiceName: "user",
	})
	require.NoError(t, err)

	addr := fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := NewUserServiceClient(conn)

	resp, err := client.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}
