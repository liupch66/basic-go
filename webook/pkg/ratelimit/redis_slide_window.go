package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaSlideWindow string

type RedisSlideWindowLimiter struct {
	cmd    redis.Cmdable
	window time.Duration
	rate   int
}

// NewRedisSlideWindowLimiter
// go 返回实际类型: NewRedisSlideWindowLimiter(cmd redis.Cmdable, rate int, window time.Duration) *RedisSlideWindowLimiter
// wire 要求返回接口类型: NewRedisSlideWindowLimiter(cmd redis.Cmdable, rate int, window time.Duration) Limiter
func NewRedisSlideWindowLimiter(cmd redis.Cmdable, rate int, window time.Duration) Limiter {
	return &RedisSlideWindowLimiter{cmd: cmd, rate: rate, window: window}
}

func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key}, r.window, r.rate, time.Now().UnixMilli()).Bool()
}
