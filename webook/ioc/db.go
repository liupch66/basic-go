package ioc

import (
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"

	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	// 读取配置法一
	// dsn := viper.GetString("db.mysql.dsn")
	// 读取配置法二
	type Config struct {
		Dsn string `yaml:"dsn"`
	}
	var cfg Config
	// 设置默认配置法一
	// viper.SetDefault("db.mysql.dsn", "root:root@tcp(localhost:3306)/webook_default")
	// 设置默认配置法二
	// var cfg Config = Config{Dsn: "root:root@tcp(localhost:3306)/webook_default"}
	// if err := viper.UnmarshalKey("db.mysql", &cfg); err != nil {
	// 	panic(err)
	// }

	// remote 不能使用 db.mysql
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}
	// db, err := gorm.Open(mysql.Open(cfg.Dsn)) // &gorm.Config{Logger: glogger.Default.LogMode(glogger.Info),},
	db, err := gorm.Open(mysql.Open(cfg.Dsn), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 一次磁盘 IO 不到 10ms,慢阈值可以设置为 50ms 或 100ms
			SlowThreshold:             time.Millisecond * 100,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			// 生产环境设置为 true,效果: INSERT INTO `users` (`email`,`password`,
			// `phone`,`wechat_open_id`,`wechat_union_id`,`ctime`,`utime`) VALUES (?,?,?,?,?,?,?)
			// ParameterizedQueries: true,
			LogLevel: glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
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
