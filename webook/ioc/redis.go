package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("redis", &cfg); err != nil {
		panic(err)
	}
	cmd := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return cmd
}

func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
