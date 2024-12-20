package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initViper() {
	// 配置命令行参数
	valPtr := pflag.String("config", "config/config.yaml", "指定配置文件")
	pflag.Parse()
	// 监听配置文件变更（蛋疼的是只知道文件变了，不知道变了啥）
	viper.WatchConfig()
	viper.SetConfigFile(*valPtr)
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("配置文件：%s，变更：%d\n", in.Name, in.Op)
	})
	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func main() {
	initViper()
	app := InitApp()
	
	for _, c := range app.consumers {
		if err := c.Start(); err != nil {
			panic(err)
		}
	}

	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}
