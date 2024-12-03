package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Article 制作库
type Article struct {
	Id    int64  `gorm:"primaryKey,Increment"`
	Title string `gorm:"type:varchar(1024)"`
	// 在 GORM 中，BLOB 是用于存储二进制数据的字段类型，通常用于存储图像、文件、加密数据等不适合存储为普通字符串的内容。
	// 在数据库中，BLOB（Binary Large Object）是一个数据类型，用来存储大块的二进制数据。
	Content  string `gorm:"type:blob"`
	AuthorId int64  `gorm:"index"`
	Ctime    int64
	Utime    int64
}

type ArticleDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{db: db}
}

func (dao *GORMArticleDAO) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}
