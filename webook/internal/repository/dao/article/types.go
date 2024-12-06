package article

import (
	"context"
	"errors"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

type ArticleDAO interface {
	// Insert 和 Update 操作制作表
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	// Upsert 操作线上表
	Upsert(ctx context.Context, art PublishedArticle) error
	// Sync 同步制作表和线上表
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
}
