package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/interact/repository/dao"
)

func InitDB() *gorm.DB {
	type Config struct {
		Dsn string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Dsn))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}
