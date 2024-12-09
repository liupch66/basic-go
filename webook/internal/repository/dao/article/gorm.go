package article

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
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
		// return fmt.Errorf("更新失败,可能是创作者非法, article_id: %d, user_id: %d", art.Id, art.AuthorId)
		return ErrPossibleIncorrectAuthor
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
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		return txDAO.Upsert(ctx, PublishedArticle(art))
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
			// return fmt.Errorf("有人非法设置别人的文章为仅自己可见, article_id: %d, author_id: %d", id, authorId)
			return ErrPossibleIncorrectAuthor
		}
		// 上面设置了 id 和 author_id 的双重验证,这里可以忽略 author_id
		res = tx.Model(&PublishedArticle{}).Where("id=?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("author_id = ?", uid).
		Offset(offset).Limit(limit).Order("utime DESC").Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).First(&art).Error
	return art, err
}
