package article

import (
	"context"

	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

// PublishedArticle 线上库,这里是组合;也可以考虑重新定义一张表
type PublishedArticle struct {
	Article
}

type GORMReaderDAO struct {
	db *gorm.DB
}

func NewGORMReaderDAO(db *gorm.DB) ReaderDAO {
	return &GORMReaderDAO{db: db}
}

func (dao *GORMReaderDAO) Upsert(ctx context.Context, art Article) error {
	// TODO implement me
	panic("implement me")
}

func (dao *GORMReaderDAO) UpsertV2(ctx context.Context, art PublishedArticle) error {
	// TODO implement me
	panic("implement me")
}
