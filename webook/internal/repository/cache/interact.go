package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/domain"
)

const (
	argReadCnt    = "read_cnt"
	argLikeCnt    = "like_cnt"
	argCollectCnt = "collect_cnt"
)

//go:embed lua/interact_incr_cnt.lua
var luaIncrCnt string

type InteractCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interact, error)
	Set(ctx context.Context, biz string, bizId int64, interact domain.Interact) error
}

type RedisInteractCache struct {
	cmd redis.Cmdable
}

func NewRedisInteractCache(cmd redis.Cmdable) InteractCache {
	return &RedisInteractCache{cmd: cmd}
}

func (cache *RedisInteractCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interact:%s:%d", biz, bizId)
}

func (cache *RedisInteractCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.cmd.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, argReadCnt, 1).Err()
}

func (cache *RedisInteractCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.cmd.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, argLikeCnt, 1).Err()
}

func (cache *RedisInteractCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.cmd.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, argLikeCnt, -1).Err()
}

func (cache *RedisInteractCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.cmd.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, argCollectCnt, 1).Err()
}

func (cache *RedisInteractCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interact, error) {
	// HMGET key field [field ...]
	// 返回一个接口{}，以区分空字符串和 nil 值
	data, err := cache.cmd.HGetAll(ctx, cache.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interact{}, err
	}
	// 还是得一个个判断，因为有可能只有阅读数没有点赞数之类的，这里偷懒没写
	if len(data) == 0 {
		return domain.Interact{}, ErrKeyNotExist
	}
	readCnt, _ := strconv.ParseInt(data[argReadCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[argLikeCnt], 10, 64)
	collectCnt, _ := strconv.ParseInt(data[argCollectCnt], 10, 64)
	return domain.Interact{
		BizId:      bizId,
		ReadCnt:    readCnt,
		LikeCnt:    likeCnt,
		CollectCnt: collectCnt,
	}, nil
}

func (cache *RedisInteractCache) Set(ctx context.Context, biz string, bizId int64, interact domain.Interact) error {
	// HMSET key field value [field value ...]
	key := cache.key(biz, bizId)
	err := cache.cmd.HMSet(ctx, key,
		argReadCnt, interact.ReadCnt,
		argLikeCnt, interact.LikeCnt,
		argCollectCnt, interact.CollectCnt,
	).Err()
	if err != nil {
		return err
	}
	return cache.cmd.Expire(ctx, key, 15*time.Minute).Err()
}
