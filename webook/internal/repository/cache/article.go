package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DeleteFirstPage(ctx context.Context, uid int64) error
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
}

type RedisArticleCache struct {
	cmd redis.Cmdable
}

func NewRedisArticleCache(cmd redis.Cmdable) ArticleCache {
	return &RedisArticleCache{cmd: cmd}
}

func (cache *RedisArticleCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (cache *RedisArticleCache) authorArtKey(uid int64) string {
	return fmt.Sprintf("article:author:%d", uid)
}

func (cache *RedisArticleCache) readerArtKey(id int64) string {
	return fmt.Sprintf("article:reader:%d", id)
}

func (cache *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	res, err := cache.cmd.Get(ctx, cache.firstPageKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(res, &arts)
	return arts, err
}

func (cache *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	// 不需要把整个 content 缓存下来,因为列表页也只展示了摘要
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, cache.firstPageKey(uid), data, 10*time.Minute).Err()
}

func (cache *RedisArticleCache) DeleteFirstPage(ctx context.Context, uid int64) error {
	return cache.cmd.Del(ctx, cache.firstPageKey(uid)).Err()
}

func (cache *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, cache.authorArtKey(art.Id), data, 10*time.Second).Err()
}

func (cache *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	res, err := cache.cmd.Get(ctx, cache.readerArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(res, &art)
	return art, err
}

func (cache *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, cache.readerArtKey(art.Id), data, 30*time.Minute).Err()
}
