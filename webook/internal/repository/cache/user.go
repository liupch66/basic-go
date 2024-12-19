package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/liupch66/basic-go/webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Set(ctx context.Context, u domain.User) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Delete(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, cache.key(u.Id), val, cache.expiration).Err()
}

func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	// 区分业务
	// ctx = context.WithValue(ctx, "biz", "user")
	// 区分子业务
	// ctx = context.WithValue(ctx, "pattern", "user:info:$id")
	val, err := cache.cmd.Get(ctx, cache.key(id)).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return cache.cmd.Del(ctx, cache.key(id)).Err()
}
