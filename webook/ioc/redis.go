package ioc

import (
	"github.com/redis/go-redis/v9"

	"basic-go/webook/config"
)

func InitRedis() redis.Cmdable {
	redisCfg := config.Config.Redis
	cmd := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	return cmd
}
