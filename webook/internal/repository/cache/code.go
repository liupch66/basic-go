package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("验证码发送太频繁")
	ErrUnknownForCode    = errors.New("发送或验证验证码遇到未知错误")
	ErrCodeVerifyTooMany = errors.New("验证次数太多")
	ErrCodeVerifyExpired = errors.New("验证码已过期")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

func (cache *RedisCodeCache) key(biz, phone string) string {
	// key 设置为 phone_code:$biz:$phone
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (cache *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.cmd.Eval(ctx, luaSetCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeSendTooMany
	default:
		return ErrUnknownForCode
	}
}

func (cache *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := cache.cmd.Eval(ctx, luaVerifyCode, []string{cache.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case -3:
		return false, ErrCodeVerifyExpired
	case 0:
		return true, nil
	case -1:
		zap.L().Warn("短信发送太频繁", zap.String("biz", biz))
		return false, ErrCodeVerifyTooMany
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForCode
	}
}
