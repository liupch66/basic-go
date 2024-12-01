package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"basic-go/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
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
	db, err := gorm.Open(mysql.Open(cfg.Dsn)) // &gorm.Config{Logger: logger.Default.LogMode(logger.Info),},
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
