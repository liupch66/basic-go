package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	"github.com/liupch66/basic-go/webook/interact/service"
	"github.com/liupch66/basic-go/webook/internal/web/client"
)

// InitInteractGRPCClient 这里灰度发布还需要引入 webook/interact/service 的代码，所以并没有把 service 建到 internal 包内
// 这是流量控制的客户端
func InitInteractGRPCClient(svc service.InteractService) interactv1.InteractServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("grpc.client.interact", &cfg); err != nil {
		panic(err)
	}

	var opts []grpc.DialOption
	if cfg.Secure {
		// 这里要加载证书什么的，启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	remote := interactv1.NewInteractServiceClient(cc)
	local := client.NewInteractLocalAdapter(svc)
	res := client.NewInteractGrayscaleRelease(remote, local, cfg.Threshold)

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		var newCfg Config
		if err := viper.UnmarshalKey("grpc.client.interact", &cfg); err != nil {
			panic(err)
		}
		res.UpdateThreshold(newCfg.Threshold)
	})

	return res
}

// InitEtcdClient 这里依赖注入，注意 server close 的时候会关掉 client（其他 service 共用 client 的时候）
func InitEtcdClient() *clientv3.Client {
	var cfg clientv3.Config
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(err)
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		panic(err)
	}
	return cli
}

// InitInteractGRPCClientV1 真正的 gRPC 客户端，这个版本是使用 ETCD 做服务注册发现的版本
func InitInteractGRPCClientV1(cli *clientv3.Client) interactv1.InteractServiceClient {
	type Config struct {
		Name   string `yaml:"name"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("grpc.client.interact", &cfg); err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(cli)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(bd)}
	if cfg.Secure {
		// 这里要加载证书什么的，启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient("etcd:///service/"+cfg.Name, opts...)
	return interactv1.NewInteractServiceClient(cc)
}
