package article

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Article 制作库
type Article struct {
	Id    int64  `gorm:"primaryKey,Increment"`
	Title string `gorm:"type:varchar(1024)"`
	// 在 GORM 中，BLOB 是用于存储二进制数据的字段类型，通常用于存储图像、文件、加密数据等不适合存储为普通字符串的内容。
	// 在数据库中，BLOB（Binary Large Object）是一个数据类型，用来存储大块的二进制数据。
	Content  string `gorm:"type:blob"`
	AuthorId int64  `gorm:"index"`
	Status   uint8
	Ctime    int64
	Utime    int64
}

type ArticleDAO interface {
	// Insert Update 操作制作表
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	// Upsert 操作线上表
	Upsert(ctx context.Context, art PublishedArticle) error
	// Sync 同步制作表和线上表
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{db: db}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) Update(ctx context.Context, art Article) error {
	art.Utime = time.Now().UnixMilli()
	// 防止修改别人的帖子, where id = ? ----> where id = ? and author_id = ?
	res := dao.db.WithContext(ctx).Model(&art).Where("author_id = ?", art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
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

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		}),
	}).Create(&art).Error
	return err
}

// Sync MySQL upsert 没有 where 条件,所以只有线上库 upsert 了,制作库并没有
func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	// 闭包形态, gorm 帮我们管理了事务的生命周期
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.Update(ctx, art)
		}
		if err != nil {
			return err
		}
		return txDAO.Upsert(ctx, PublishedArticle{Article: art})
	})
	return id, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	// 小优化,尽量减少事务时间
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id=? AND author_id=?", id, authorId).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("有人非法设置别人的文章为仅自己可见, article_id: %d, author_id: %s", id, authorId)
		}
		// 上面设置了 id 和 author_id 的双重验证,这里可以忽略 author_id
		return tx.Model(&PublishedArticle{}).Where("id=?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
	})
}
