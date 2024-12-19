package dao

import (
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/internal/repository/dao/article"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&AsyncSms{},
		&CronJob{},
	)
}
