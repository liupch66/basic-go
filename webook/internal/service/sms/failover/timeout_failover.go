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
	// 连续超时次数
	cnt int32
	// 连续超时次数阈值
	threshold int32
}

func NewServiceForTimeout(svcs []sms.Service, threshold int32) *ServiceForTimeout {
	return &ServiceForTimeout{svcs: svcs, threshold: threshold}
}

func (s *ServiceForTimeout) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	// 严格的轮询,不管发送成功还是达到超时次数阈值,下一次都切换服务商
	length := int32(len(s.svcs))
	// 尝试切换服务商的次数
	attempts := int32(0)
	for attempts < length {
		// idx 能及时更新
		idx := atomic.LoadInt32(&s.idx)
		// 获取当前服务商
		svc := s.svcs[idx%length]
		err := svc.Send(ctx, tplId, params, numbers...)
		if err == nil {
			newIdx := (idx + 1) % length
			// 使用 CAS 确保只会有一个 goroutine 能够切换到下一个服务商
			if atomic.CompareAndSwapInt32(&s.idx, idx, newIdx) {
				atomic.StoreInt32(&s.cnt, 0)
			}
			return nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			// 几个 goroutine 并发失败这里就加几,所以这里不用 CAS
			atomic.AddInt32(&s.cnt, 1)
			if atomic.LoadInt32(&s.cnt) >= s.threshold {
				newIdx := (idx + 1) % length
				if atomic.CompareAndSwapInt32(&s.idx, idx, newIdx) {
					atomic.StoreInt32(&s.cnt, 0)
					attempts++
				}
			}
		}
	}
	return errors.New("所有短信服务商都发送失败")
}

// Share memory by communicating; don't communicate by sharing memory.
// 通过通信共享内存；不要通过共享内存进行通信。
