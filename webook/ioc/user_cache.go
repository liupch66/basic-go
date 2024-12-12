package ioc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/pkg/redisx"
)

// InitUserCache 配合 PrometheusHook 使用。这里只能传 *redis.Client，因为 redis.Cmdable 没有 AddHook 方法
// 而且 initRedis 也要改成生成 *redis.Client 而不是 redis.Cmdable。还要 InitArticleCache...，很麻烦
func InitUserCache(client *redis.Client) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(prometheus.SummaryOpts{
		Namespace: "geektime",
		Subsystem: "webook",
		Name:      "redis_resp_time",
		Help:      "统计 redis 耗时和缓存命中率",
		ConstLabels: map[string]string{
			"biz": "user",
		},
	}))
	panic("别调用")
}
