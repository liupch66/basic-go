package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initViper()
	initLogger()
	fmt.Println(viper.AllKeys())
	fmt.Println(viper.AllSettings())
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, world!")
	})
	err := server.Run("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
}

func initViper() {
	// // 文件名(不包含拓展名)
	// viper.SetConfigName("dev")
	// viper.SetConfigType("yaml")
	// // 当前工作目录下的 config
	// // 可以多次添加搜索路径
	// viper.AddConfigPath("./config")
	// // 搜索,读取文件
	// if err := viper.ReadInConfig(); err != nil {
	// 	panic(err)
	// }

	viper.SetConfigFile("webook/config/dev.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViperV1() {
	viper.SetConfigType("yaml")
	cfg := `
db.mysql:
  dsn: "root:root@tcp(localhost:3306)/webook"
redis:
  addr: "localhost:6379"
  password: ""
  DB: 1
`
	err := viper.ReadConfig(strings.NewReader(cfg))
	if err != nil {
		panic(err)
	}
}

// 启动参数
func initViperV2() {
	s := pflag.String("config", "webook/config/config.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*s)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:22379", "/webook")
	if err != nil {
		panic(err)
	}
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	// 这里没用, Viper 默认并不会自动监听远程配置的变化
	// viper.OnConfigChange(func(in fsnotify.Event) {
	// 	fmt.Println(in.Name, in.Op)
	// })
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initViperV3() {
	viper.SetConfigFile("webook/config/dev.yaml")
	// 监听文件变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 这里只知道变了,不知道哪里变了
		fmt.Println(in.Name, in.Op)
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.L().Info("before replace")
	zap.ReplaceGlobals(logger)
	zap.L().Info("some info")
	type demo struct {
		Name string `json:"name"`
	}
	zap.L().Info("这是实验参数", zap.Error(errors.New("an error")), zap.Int64("id", 64), zap.Any("一个结构体", demo{Name: "daming"}))
}
