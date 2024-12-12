package redisx

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	summaryVec *prometheus.SummaryVec
}

func NewPrometheusHook(opts prometheus.SummaryOpts) *PrometheusHook {
	// 可以考虑加 biz，但是比较麻烦，每个地方都要加
	summaryVec := prometheus.NewSummaryVec(opts, []string{"cmd", "key_exists"})
	prometheus.MustRegister(summaryVec)
	return &PrometheusHook{summaryVec: summaryVec}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 相当于啥也没干
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		startTime := time.Now()
		var err error
		defer func() {
			// biz := ctx.Value("biz")
			duration := time.Since(startTime).Milliseconds()
			// Nil reply returned by Redis when key does not exist.
			keyExists := errors.Is(err, redis.Nil)
			// 也可以考虑 cmd.FullName()
			p.summaryVec.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	// TODO implement me
	panic("implement me")
}
