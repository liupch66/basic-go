package startup

import (
	"context"
	"database/sql"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"basic-go/webook/internal/repository/dao"
)

var db *gorm.DB

func InitTestDB() *gorm.DB {
	if db == nil {
		// 检查是否已有数据库连接实例
		dsn := "root:root@tcp(localhost:3306)/webook"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		// 通过 PingContext 检查数据库连接是否正常
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Println("等待连接 MySQL", err)
		}
		// 一旦数据库连接正常，使用 GORM 打开 MySQL 数据库连接
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		// 执行数据库表的初始化
		err = dao.InitTables(db)
		if err != nil {
			panic(err)
		}
	}
	return db
}
