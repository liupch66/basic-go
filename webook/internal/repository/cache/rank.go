package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/domain"
)

type RankCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RedisRankCache struct {
	cmd redis.Cmdable
	key string
}

func NewRedisRankCache(cmd redis.Cmdable) *RedisRankCache {
	return &RedisRankCache{cmd: cmd, key: "rank"}
}

func (cache *RedisRankCache) Set(ctx context.Context, arts []domain.Article) error {
	// 热度榜缓存，不缓存内容，保证缓存内容能返回查询接口所需所有数据
	for _, art := range arts {
		art.Content = ""
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 过期时间尽量长点，保证热度榜计算出来（包括重试时间）
	// 也可以考虑设置永不过期，防止热度榜计算出错时，这个缓存过期后没有热度榜。反正后续 set 也能更新热度榜
	// 过期时间也可以设置成字段
	return cache.cmd.Set(ctx, cache.key, data, 10*time.Minute).Err()
	// 也可以提前把热度榜的 articles 写到缓存里，按 id => article 进行映射好
}

func (cache *RedisRankCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := cache.cmd.Get(ctx, cache.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}
