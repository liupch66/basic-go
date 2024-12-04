package dao

import (
	"context"
	"fmt"
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
	Update(ctx context.Context, art Article) error
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

func (dao *GORMArticleDAO) Update(ctx context.Context, art Article) error {
	art.Utime = time.Now().UnixMilli()
	// Updates 使用 struct 或 map[string]interface{}.用 struct 时，默认情况下 GORM 只会更新非零值的字段. map 会更新零值
	// 下面两个写法效果相同,可读性较差(不知道更新了什么)
	// return dao.db.WithContext(ctx).Model(&Article{}).Where("id=?", art.Id).Updates(art).Error
	// return dao.db.WithContext(ctx).Model(&art).Updates(art).Error
	// 使用 map 可读性较好
	// 防止修改别人的帖子, where id = ? ----> where id = ? and author_id = ?
	res := dao.db.WithContext(ctx).Model(&art).Where("author_id = ?", art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败,可能是创作者非法, article_id: %d, user_id: %d", art.Id, art.AuthorId)
	}
	return nil
}
