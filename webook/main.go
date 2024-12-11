package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initViper()
	initLogger()
	zap.L().Info("测试", zap.Any("准备：", "OK"))
	app := InitApp()

	// 坑，这里不能放到 server.Run 后面，不然没用
	for _, consumer := range app.consumers {
		err := consumer.Start()
		if err != nil {
			// 这里实现并不好，前面的 consumer 失败会导致后面无法运行，可以考虑回滚
			panic(err)
		}
	}

	server := app.web
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, world!")
	})
	server.Run(":8080")
}

func initViper() {
	viper.SetConfigFile("webook/config/dev.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
