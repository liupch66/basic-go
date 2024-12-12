package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"basic-go/webook/internal/service/sms"
)

type PrometheusDecorator struct {
	svc        sms.Service
	summaryVec *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geektime",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 的性能数据",
	}, []string{"tplId"})
	prometheus.MustRegister(summaryVec)
	return &PrometheusDecorator{svc: svc, summaryVec: summaryVec}
}

func NewPrometheusDecoratorV1(opts prometheus.SummaryOpts, labelNames []string) *PrometheusDecorator {
	summaryVec := prometheus.NewSummaryVec(opts, labelNames)
	prometheus.MustRegister(summaryVec)
	return &PrometheusDecorator{summaryVec: summaryVec}
}

func (p *PrometheusDecorator) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.summaryVec.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, tplId, params, numbers...)
}
