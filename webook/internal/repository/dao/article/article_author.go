package article

import (
	"context"

	"gorm.io/gorm"
)

type AuthorDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
}

type GORMAuthorDAO struct {
	db *gorm.DB
}

func NewGORMAuthorDAO(db *gorm.DB) AuthorDao {
	return &GORMAuthorDAO{db: db}
}

func (dao *GORMAuthorDAO) Insert(ctx context.Context, art Article) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (dao *GORMAuthorDAO) Update(ctx context.Context, art Article) error {
	// TODO implement me
	panic("implement me")
}
