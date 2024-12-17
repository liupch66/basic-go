package job

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"basic-go/webook/pkg/logger"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vec    *prometheus.SummaryVec
	tracer trace.Tracer
}

func NewCronJobBuilder(l logger.LoggerV1) *CronJobBuilder {
	vec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geektime",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "统计定时任务的执行情况",
	}, []string{"name", "success"})
	prometheus.MustRegister(vec)
	return &CronJobBuilder{l: l, vec: vec, tracer: otel.GetTracerProvider().Tracer("webook/internal/job")}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		_, span := b.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		b.l.Info("任务开始", logger.String("job", name))
		var success bool
		defer func() {
			b.l.Info("任务结束", logger.String("job", name))
			duration := time.Since(start).Milliseconds()
			b.vec.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run()
		success = err == nil
		if err != nil {
			span.RecordError(err)
			b.l.Error("运行任务失败", logger.Error(err), logger.String("job", name))
			return err
		}
		return nil
	})
}

type cronJobFuncAdapter func() error

func (c cronJobFuncAdapter) Run() {
	_ = c()
}
