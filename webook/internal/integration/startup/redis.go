package startup

import (
	"context"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

var redisClient redis.Cmdable

func InitRedis() redis.Cmdable {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		for err := redisClient.Ping(context.Background()).Err(); err != nil; {
			panic(err)
		}
	}
	return redisClient
}

func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
