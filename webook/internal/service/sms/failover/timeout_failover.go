package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"basic-go/webook/internal/service/sms"
)

type ServiceForTimeout struct {
	svcs []sms.Service
	idx  int32
	// 连续超时次数 cnt
	cnt int32
	// 连续超时阈值 threshold
	threshold int32
}

func (s *ServiceForTimeout) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	idx := atomic.LoadInt32(&s.idx)
	cnt := atomic.LoadInt32(&s.cnt)
	if cnt > s.threshold {
		newIdx := (idx + 1) % int32(len(s.svcs))
		if atomic.CompareAndSwapInt32(&s.idx, idx, newIdx) {
			// 切换新的服务商成功,重置计数器
			atomic.StoreInt32(&s.cnt, 0)
		}
		// CAS 操作失败,说明有人切换了
		idx = newIdx
	}
	svc := s.svcs[idx]
	err := svc.Send(ctx, tplId, params, numbers...)
	switch {
	case err == nil:
		// 发送成功,重置计数器
		atomic.StoreInt32(&s.cnt, 0)
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&s.cnt, 1)
	default:
		// 其他错误,保持不动
	}
	return err
}
