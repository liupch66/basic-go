package opentelemetry

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TestServer 测试 Gin Web 服务器，配置并启动 OpenTelemetry 跟踪，进行分布式追踪测试
func TestServer(t *testing.T) {
	// 初始化服务资源信息，传递服务名和版本信息
	res, err := newResource("demo", "v0.0.1")
	require.NoError(t, err)

	// 初始化用于跨进程传播追踪信息的 propagator
	prop := newPropagator()
	// 设置 OpenTelemetry 使用的 TextMapPropagator，允许跨进程传递 trace context
	otel.SetTextMapPropagator(prop)

	// 初始化 TraceProvider，用于管理和导出追踪信息
	tp, err := newTraceProvider(res)
	require.NoError(t, err)
	// 在测试结束时关闭 TraceProvider，确保批量数据被导出
	defer tp.Shutdown(context.Background())
	// 设置全局的 TraceProvider，OpenTelemetry 在后续的代码中会使用该 provider
	otel.SetTracerProvider(tp)

	// 创建一个新的 Gin 路由
	server := gin.Default()

	// 设置一个测试路由，处理 GET 请求
	server.GET("/test", func(ginCtx *gin.Context) {
		// 创建一个 Tracer，Tracer 的名字最好是唯一的，通常用当前包名
		tracer := otel.Tracer("opentelemetry")

		// 获取当前 Gin 请求的 context
		var ctx context.Context = ginCtx
		// 创建一个新的 Span，表示一个追踪的开始
		ctx, span := tracer.Start(ctx, "top-span")
		// 确保在函数返回时结束该 Span
		defer span.End()

		// 在当前 Span 中添加事件，便于后续查看和分析
		span.AddEvent("event-1")

		// 模拟一些耗时操作
		time.Sleep(time.Second)

		// 创建一个子 Span，表示当前 Span 下的一个操作
		ctx, subSpan := tracer.Start(ctx, "sub-span")
		// 确保在函数返回时结束子 Span
		defer subSpan.End()
		// 模拟一些耗时操作
		time.Sleep(time.Millisecond * 300)

		// 设置子 Span 的属性，增加追踪信息
		subSpan.SetAttributes(attribute.String("key1", "value1"))

		// 返回响应
		ginCtx.String(http.StatusOK, "OK")
	})

	// 启动服务器并监听请求
	server.Run(":8083")
}

// newResource 创建 OpenTelemetry 资源对象，包含服务名和版本等信息
func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	// 合并默认资源与自定义的资源信息
	return resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL,
		// 设置服务的名称和版本信息
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	))
}

// newPropagator 创建一个组合型的 propagator，支持 TraceContext 和 Baggage 的传播
func newPropagator() propagation.TextMapPropagator {
	// 使用 CompositeTextMapPropagator 来组合 TraceContext 和 Baggage
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

// newTraceProvider 创建一个新的 TraceProvider，用于导出追踪数据
func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	// 配置 Zipkin 导出器，将数据发送到 Zipkin 收集器
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	// 创建一个 TraceProvider，指定导出器和资源信息
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)), // 配置批量导出，超时时间为 1 秒
		trace.WithResource(res), // 设置服务资源信息
	)
	return traceProvider, nil
}
