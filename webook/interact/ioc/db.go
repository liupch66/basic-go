package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	"github.com/liupch66/basic-go/webook/interact/repository/dao"
	prom "github.com/liupch66/basic-go/webook/pkg/gormx/callback/prometheus"
	"github.com/liupch66/basic-go/webook/pkg/gormx/connpool"
)

// 这里都是返回 *gorm.DB, wire 不好处理这种返回同类型的，稍微处理一下，不用 wire 的话就不用处理
type (
	SrcDB *gorm.DB
	DstDB *gorm.DB
)

func InitSrcDB() SrcDB {
	return initDB("db.src", "webook")
}

func InitDstDB() DstDB {
	return initDB("db.dst", "webook_interact")
}

func InitDoubleWritePool(srcDB SrcDB, dstDB DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	pool := connpool.NewDoubleWritePool(srcDB, dstDB, pattern)
	// todo: 怎么持续监听配置变更？
	// go func() {
	// 	for {
	// 		viper.WatchConfig()
	// 		viper.OnConfigChange(func(in fsnotify.Event) {
	// 			pattern := viper.GetString("migrator.pattern")
	// 			pool.UpdatePattern(pattern)
	// 		})
	// 	}
	// }()
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		pattern := viper.GetString("migrator.pattern")
		pool.UpdatePattern(pattern)
	})
	return pool
}

// InitBizDB 这个是业务用的，支持双写的 DB
func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: pool,
	}))
	if err != nil {
		panic(err)
	}
	return db
}

// 这里可以修改为私有
func initDB(key, name string) *gorm.DB {
	type Config struct {
		Dsn string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey(key, &cfg); err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Dsn))
	if err != nil {
		panic(err)
	}

	// 接入 prometheus
	err = db.Use(prometheus.New(prometheus.Config{
		DBName: name,
		// 刷新间隔时间，单位秒，定义多长时间刷新一次数据，默认值就是 15s
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{&prometheus.MySQL{
			// 设置 MySQL 变量 'Threads_running' 作为监控指标
			VariableNames: []string{"Threads_running"},
		}},
	}))
	if err != nil {
		panic(err)
	}

	// 可以从这个例子学习怎么实现 gorm.Plugin 接口和 gorm 的 hook 结合
	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics()))
	if err != nil {
		panic(err)
	}

	cb := prom.Callbacks{
		Namespace:  "geektime",
		Subsystem:  "webook",
		Name:       "gorm_" + name,
		InstanceID: "my-instance-1",
		Help:       "gorm DB 查询",
	}
	err = cb.Register(db)
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
