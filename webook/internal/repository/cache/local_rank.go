package cache

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/syncx/atomicx"

	"github.com/liupch66/basic-go/webook/internal/domain"
)

type LocalRankCache struct {
	// 可以考虑直接使用 uber 的，或者 SDK 自带的
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewLocalRankCache() *LocalRankCache {
	return &LocalRankCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf(time.Now()),
		expiration: 10 * time.Minute,
	}
}

func (r *LocalRankCache) Set(ctx context.Context, arts []domain.Article) error {
	// 可能会有并发问题，因为是两个原子操作。解决：原子操作下面的 item
	r.topN.Store(arts)
	ddl := time.Now().Add(r.expiration)
	r.ddl.Store(ddl)
	return nil
	// 也可以按照 id => article 提前缓存好
}

func (r *LocalRankCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	arts := r.topN.Load()
	// 初始化后 arts 是空的，这里做个判断
	if ddl.Before(time.Now()) || len(arts) == 0 {
		return nil, errors.New("本地缓存未命中")
	}
	return arts, nil
}

func (r *LocalRankCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	return arts, nil
}

type item struct {
	arts []domain.Article
	ddl  time.Time
}
