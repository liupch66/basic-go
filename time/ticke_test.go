package time

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	tm := time.NewTimer(time.Second)
	defer tm.Stop()
	now := <-tm.C
	t.Log(now.Unix())
}

func TestTicker(t *testing.T) {
	tc := time.NewTicker(time.Second)
	defer tc.Stop()
	// for now := range tc.C {
	// 	t.Log(now.Unix())
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// break end 时 end 只能在 break 前面，goto end 时 end 可前可后
	// break end 时 end 后面不能有其他操作，它直接跳出循环，goto end 时 end 后面可以有任意操作，它跳转到指定的标签处继续执行
end:
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Log("超时")
			} else if errors.Is(ctx.Err(), context.Canceled) {
				t.Log("取消")
			}
			break end
		case now := <-tc.C:
			t.Log(now.Unix())
		}
	}
}
