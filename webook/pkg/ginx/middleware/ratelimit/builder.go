package ratelimit

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/liupch66/basic-go/webook/pkg/ratelimit"
)

//go:embed slide_window.lua
var slideWindowLuaScript string

type Builder struct {
	cmd      redis.Cmdable
	prefix   string
	interval time.Duration
	rate     int
}

func NewBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *Builder {
	return &Builder{
		cmd:      cmd,
		prefix:   "ip-limit",
		interval: interval,
		rate:     rate,
	}
}

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) ipLimit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.cmd.Eval(ctx, slideWindowLuaScript, []string{key}, b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.ipLimit(ctx)
		if !errors.Is(err, redis.Nil) {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 坑！！！！！！这里返回的 err == redis.Nil == RedisError("redis: nil") , 而 != nil
		// if err != nil {
		// 	ctx.AbortWithStatus(http.StatusInternalServerError)
		// 	return
		// }
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}

// =================================================================================================================

type BuilderV1 struct {
	prefix  string
	limiter ratelimit.Limiter
}

func NewBuilderV1(prefix string, limiter ratelimit.Limiter) *BuilderV1 {
	return &BuilderV1{prefix: prefix, limiter: limiter}
}

func (b *BuilderV1) Limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limiter.Limit(ctx, key)
}
