package ioc

import (
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	dao2 "basic-go/webook/interact/repository/dao"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		Dsn string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.Dsn), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold:             time.Millisecond * 100,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
		Labels: nil,
	}))
	if err != nil {
		panic(err)
	}

	// c := newCallbacks()
	// c.registerAll(db)
	err = db.Use(newCallbacks())
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithDBName("webook"), tracing.WithoutQueryVariables()))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

// Printf 单方法的接口可以这样使用
func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

type callbacks struct {
	vector *prom.SummaryVec
}

func newCallbacks() *callbacks {
	vector := prom.NewSummaryVec(prom.SummaryOpts{
		Namespace:   "geektime",
		Subsystem:   "webook",
		Name:        "gorm_query_time",
		Help:        "统计 GORM 的执行时间",
		ConstLabels: map[string]string{"db": "webook"},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	prom.MustRegister(vector)
	return &callbacks{vector: vector}
}

func (c *callbacks) Before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (c *callbacks) After(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		start, ok := val.(time.Time)
		if !ok {
			// 啥也干不了，最多记日志
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(start).Milliseconds()))
	}
}

func (c *callbacks) registerAll(db *gorm.DB) {
	err := db.Callback().Create().Before("*").Register("prometheus_create_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").Register("prometheus_create_after", c.After("create"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().Before("*").Register("prometheus_update_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").Register("prometheus_update_after", c.After("update"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().Before("*").Register("prometheus_query_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").Register("prometheus_query_after", c.After("query"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", c.After("delete"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Raw().After("*").Register("prometheus_raw_after", c.After("raw"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().Before("*").Register("prometheus_row_before", c.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").Register("prometheus_row_after", c.After("row"))
	if err != nil {
		panic(err)
	}
}

func (c *callbacks) Name() string {
	return "prometheus-query"
}

func (c *callbacks) Initialize(db *gorm.DB) error {
	c.registerAll(db)
	return nil
}
