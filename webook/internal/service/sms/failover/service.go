package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"basic-go/webook/internal/service/sms"
)

type Service struct {
	svcs []sms.Service
}

func NewService(svcs []sms.Service) *Service {
	return &Service{svcs: svcs}
}

// Send 每次都从头开始轮询，绝大多数请求会在 svcs[0] 就成功，负载不均衡。
func (s *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	for _, svc := range s.svcs {
		err := svc.Send(ctx, tplId, params, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("所有短信服务商都发送失败")
}

// =================================================================================================

type ServiceV1 struct {
	svcs []sms.Service
	idx  uint64
}

func (s *ServiceV1) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	idx := atomic.AddUint64(&s.idx, 1)
	length := uint64(len(s.svcs))
	for i := idx; i < idx+length; i++ {
		// 注意取余
		err := s.svcs[i%length].Send(ctx, tplId, params, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return err
		default:
			// 输出日志
		}
	}
	return errors.New("所有短信服务商都发送失败")
}
