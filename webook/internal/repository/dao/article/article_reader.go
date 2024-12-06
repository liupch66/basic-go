package article

import (
	"context"

	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
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
