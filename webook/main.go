package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"

	"github.com/liupch66/basic-go/webook/ioc"
)

func main() {
	initViper()
	initLogger()
	initPrometheus()
	closeFunc := ioc.InitOtel()

	app := InitApp()
	for _, consumer := range app.consumers {
		err := consumer.Start()
		if err != nil {
			// 这里实现并不好，前面的 consumer 失败会导致后面无法运行，可以考虑回滚
			panic(err)
		}
	}

	cron := app.cron
	cron.Start()

	server := app.web
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, world!")
	})
	server.Run(":8080")

	// gin是封装的http.server，但是 gin 没有优雅退出，http.server 有
	// http.Server{}.Shutdown()

	// httpServer := http.Server{
	// 	Port:                         ":8080",
	// 	Handler:                     app.web.Handler(),
	// }
	// httpServer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)

	// 可以考虑超时强制退出，防止有些任务执行很长时间
	tm := time.NewTimer(10 * time.Minute)
	select {
	case <-tm.C:
	case <-cron.Stop().Done():
	}

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

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()
}
