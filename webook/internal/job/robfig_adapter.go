package job

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/liupch66/basic-go/webook/pkg/logger"
)

type RankJobAdapter struct {
	j   Job
	l   logger.LoggerV1
	sum prometheus.Summary
}

func NewRankJobAdapter(j Job, l logger.LoggerV1) *RankJobAdapter {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        "corn_job",
		ConstLabels: map[string]string{"name": j.Name()},
	})
	prometheus.MustRegister(sum)
	return &RankJobAdapter{j: j, l: l, sum: sum}
}

func (r *RankJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		r.sum.Observe(float64(duration))
	}()
	err := r.j.Run()
	if err != nil {
		r.l.Error("运行任务失败: ", logger.Error(err), logger.String("job", r.j.Name()))
	}
}
