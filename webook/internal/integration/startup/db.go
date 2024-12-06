package startup

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

var mongoDB *mongo.Database

func InitMongoDB() *mongo.Database {
	if mongoDB == nil {
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
		client, err := mongo.Connect(opts)
		if err != nil {
			panic(err)
		}
		mongoDB = client.Database("webook")
	}
	return mongoDB
}
