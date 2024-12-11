package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	// 这一个实例名字，你可以考虑使用 本地 IP，又或者在启动的时候配置一个 ID
	InstanceID string
}

func (p *PrometheusBuilder) Build() gin.HandlerFunc {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		// 下面这三个不能用 "-"，只能用 "_"
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_resp_time",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"method", "pattern", "status"})
	prometheus.MustRegister(summaryVec)

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: p.Namespace,
		Subsystem: p.Subsystem,
		Name:      p.Name + "_active_req",
		Help:      p.Help,
		ConstLabels: map[string]string{
			"instance_id": p.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)

	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()
		// 最后再统计一下执行时间
		defer func() {
			gauge.Dec()
			// ctx.FullPath()：返回的是路由匹配时使用的 路径模板，包括参数占位符。如 /users/:id
			// ctx.Request.URL.Path：返回的是客户端实际请求的路径。如 /users/123
			pattern := ctx.FullPath()
			// 考虑到 404
			if pattern == "" {
				pattern = "unknown"
			}
			summaryVec.WithLabelValues(ctx.Request.Method, pattern, strconv.Itoa(ctx.Writer.Status())).
				Observe(float64(time.Since(start)))
		}()
		ctx.Next()
	}
}
