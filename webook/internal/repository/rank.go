package repository

import (
	"context"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository/cache"
)

type RankRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankRepository struct {
	// 接口也可以
	// cache cache.RankCache

	// 这里加入本地缓存，换成具体实现，可读性强点，但对测试不友好，因为没有面向接口编程
	local *cache.RankLocalCache
	redis *cache.RedisRankCache
}

func NewCachedRankRepository(local *cache.RankLocalCache, redis *cache.RedisRankCache) RankRepository {
	return &CachedRankRepository{local: local, redis: redis}
}

func (repo *CachedRankRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = repo.local.Set(ctx, arts)
	return repo.redis.Set(ctx, arts)
}

func (repo *CachedRankRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := repo.local.Get(ctx)
	if err == nil {
		return arts, nil
	}

	arts, err = repo.redis.Get(ctx)
	// 回写本地缓存
	if err == nil {
		_ = repo.local.Set(ctx, arts)
	} else {
		// 如果 Redis 崩了，这里设置一个本地缓存兜底方案，实际意义不大
		return repo.local.ForceGet(ctx)
	}
	return arts, err
}
