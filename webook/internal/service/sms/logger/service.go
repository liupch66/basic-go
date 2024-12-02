package logger

import (
	"context"

	"go.uber.org/zap"

	"basic-go/webook/internal/service/sms"
)

type Service struct {
	svc sms.Service
}

// Send 不想每个地方都去打,可以考虑装饰器打印日志
func (s Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	zap.L().Debug("发送短信", zap.String("tplId", tplId), zap.Any("params", params))
	err := s.svc.Send(ctx, tplId, params, numbers...)
	if err != nil {
		zap.L().Error("发送短信出错", zap.Error(err))
	}
	return err
}
