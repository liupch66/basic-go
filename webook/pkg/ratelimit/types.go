package ratelimit

import (
	"context"
)

type Limiter interface {
	// Limit key 就是限流对象
	Limit(ctx context.Context, key string) (bool, error)
}
