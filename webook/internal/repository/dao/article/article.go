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
	// 如果冲突发生，会根据 OnConflict 子句中的规则更新已存在的记录，而不是插入新记录
	// 这比单独处理插入失败后的回滚逻辑更加高效
	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		/* SQL 2003 标准 ~
		INSERT INTO published_articles (id, title, content, author_id, status, ctime, utime)VALUES (...)
		ON CONFLICT (id)
		DO UPDATE SET title=x, content=x, status=x, utime=x WHERE XXX / DO NOTHING

		MySQL 最终语句:
		INSERT INTO published_articles (...) VALUES (...)
		ON DUPLICATE KEY
		UPDATE title=x...
		*/

		// 指定哪些列应该用于冲突检查, 默认情况检查主键冲突和唯一索引冲突
		// Columns: nil,
		// 当发生冲突时，并且额外符合的附加条件
		// Where: clause.Where{},
		// 指定冲突时的约束条件名称
		// OnConstraint: "",
		// 在冲突时是否执行 "什么都不做" 的操作
		// DoNothing: false,
		// 指定冲突时需要更新的列及其值, MySQL 只需要关心这里
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
		art.Id = id
		return txDAO.Upsert(ctx, PublishedArticle{Article: art})
	})
	return id, err
}
