package ioc

import (
	"github.com/IBM/sarama"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/liupch66/basic-go/webook/interact/repository/dao"
	"github.com/liupch66/basic-go/webook/pkg/ginx"
	"github.com/liupch66/basic-go/webook/pkg/gormx/connpool"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events/fixer"
	"github.com/liupch66/basic-go/webook/pkg/migrator/scheduler"
)

const topic = "migrator_interact"

func InitMigratorWeb(src SrcDB, dst DstDB, l logger.LoggerV1,
	producer events.Producer, pool *connpool.DoubleWritePool) *ginx.Server {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geektime",
		Subsystem: "webook_interact_admin",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	pattern := viper.GetString("migrator.pattern")
	// 在这里，有多少张表，就初始化多少个 scheduler
	interSch := scheduler.NewScheduler[dao.Interact](src, dst, l, pattern, producer, pool)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		pattern := viper.GetString("migrator.pattern")
		// 这里 pool 没导出，没法 interSch.pool.UpdatePattern(pattern)
		// 要不重写一个方法，要不就把 pool 导出。这里是重写的方法
		interSch.UpdatePattern(pattern)
	})
	engine := gin.Default()
	interSch.RegisterRoutes(engine.Group("/migrator"))
	// 这里可以就用一个了，多个就跟下面例子一样后面再接路由区分，
	// interSch.RegisterRoutes(engine.Group("/migrator/interact"))
	addr := viper.GetString("migrator.http.addr")
	return ginx.NewServer(engine, addr)
}

func InitMigratorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitFixDataConsumer(client sarama.Client, l logger.LoggerV1, src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interact] {
	consumer, err := fixer.NewConsumer[dao.Interact](client, l, src, dst, topic)
	if err != nil {
		panic(err)
	}
	return consumer
}
