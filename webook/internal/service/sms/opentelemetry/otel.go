package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"basic-go/webook/internal/service/sms"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("webook/internal/service/sms/opentelemetry/otel.go")
	return &Service{svc: svc, tracer: tracer}
}

func (s *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	// 我是一个调用短信服务的客户端
	ctx, span := s.tracer.Start(ctx, "sms_"+tplId, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End(trace.WithStackTrace(true))
	err := s.svc.Send(ctx, tplId, params, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
