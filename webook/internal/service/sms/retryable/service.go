package retryable

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/liupch66/basic-go/webook/internal/service/sms"
)

type Service struct {
	svc      sms.Service
	retryMax int
}

func NewService(svc sms.Service, retryMax int) sms.Service {
	return &Service{svc: svc, retryMax: retryMax}
}

func (s *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	var err error
	for i := 0; i < s.retryMax; i++ {
		err = s.svc.Send(ctx, tplId, params, numbers...)
		if err == nil {
			return nil
		}
		// 稍微增加延迟,不要立马重试
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, aborting retries")
			return ctx.Err()
		case <-time.After(time.Second):
			log.Println("Retrying after 1 second")
		}
	}
	return fmt.Errorf("短信发送重试都失败了: %w", err)
}
