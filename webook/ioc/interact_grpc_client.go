package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	"github.com/liupch66/basic-go/webook/interact/service"
	"github.com/liupch66/basic-go/webook/internal/web/client"
)

// InitInteractGRPCClient 这里灰度发布还需要引入 webook/interact/service 的代码，所以并没有把 service 建到 internal 包内
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

	viper.OnConfigChange(func(in fsnotify.Event) {
		var newCfg Config
		if err := viper.UnmarshalKey("grpc.client.interact", &cfg); err != nil {
			panic(err)
		}
		res.UpdateThreshold(newCfg.Threshold)
	})

	return res
}
