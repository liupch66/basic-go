package failover

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/service/sms"
	smsmocks "basic-go/webook/internal/service/sms/mocks"
)

func TestServiceForTimeout_Send(t *testing.T) {
	// 模拟三个服务商测试: svc0, svc1, svc2,这里测试阈值为 3
	// 这里要控制一开始的 idx 和 cnt,不然测试用例会互相影响
	// 严格的轮询,不管发送成功还是达到超时次数阈值,下一次都切换服务商
	testCases := []struct {
		name string

		mock      func(ctrl *gomock.Controller) []sms.Service
		idx       int32
		cnt       int32
		threshold int32

		expectedErr error
		expectedIdx int32
		expectedCnt int32
	}{
		// 从 svc0 开始轮询(idx = 0, cnt = 0)
		{
			name: "svc0 第一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 第一次发送超时-->第二次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 前两次发送超时,第三次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(2)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 发送三次失败--> svc1 第一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1, nil}
			},
			idx:         0,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "所有(3 个)服务商都发送失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				return []sms.Service{svc0, svc1, svc2}
			},
			idx:         0,
			cnt:         0,
			threshold:   3,
			expectedErr: errors.New("所有短信服务商都发送失败"),
			expectedIdx: 0,
			expectedCnt: 0,
		},
		// 从 svc1 开始轮询(idx = 1, cnt = 0)
		{
			name: "svc1 第一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{nil, svc1, nil}
			},
			idx:         1,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "svc1 第一次发送超时-->第二次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{nil, svc1, nil}
			},
			idx:         1,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "svc1 前两次发送超时,第三次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(2)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{nil, svc1, nil}
			},
			idx:         1,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "svc1 发送三次(阈值)失败,切换下一个服务商--> svc2 第一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{nil, svc1, svc2}
			},
			idx:         1,
			cnt:         0,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 0,
			expectedCnt: 0,
		},
		{
			name: "所有(3 个)服务商都发送失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc0 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				return []sms.Service{svc0, svc1, svc2}
			},
			idx:         1,
			cnt:         0,
			threshold:   3,
			expectedErr: errors.New("所有短信服务商都发送失败"),
			expectedIdx: 1,
			expectedCnt: 0,
		},
		// 从 svc0 开始轮询,第一次已经发送失败(idx = 0, cnt = 1)
		{
			name: "svc0 第二次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         1,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 第二次发送失败,第三次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         1,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 第二次和第三次发送失败-->svc1 第一次成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(2)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1, nil}
			},
			idx:         0,
			cnt:         1,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "所有(3 个)服务商都发送失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(2)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				return []sms.Service{svc0, svc1, svc2}
			},
			idx:         0,
			cnt:         1,
			threshold:   3,
			expectedErr: errors.New("所有短信服务商都发送失败"),
			expectedIdx: 0,
			expectedCnt: 0,
		},
		// 从 svc0 开始轮询,第一次和第二次已经发送失败(idx = 0, cnt = 2)
		{
			name: "svc0 第三次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, nil, nil}
			},
			idx:         0,
			cnt:         2,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 1,
			expectedCnt: 0,
		},
		{
			name: "svc0 第三次发送失败--> svc1 第一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1, nil}
			},
			idx:         0,
			cnt:         2,
			threshold:   3,
			expectedErr: nil,
			expectedIdx: 2,
			expectedCnt: 0,
		},
		{
			name: "所有(3 个)服务商都发送失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc2 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(1)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).Times(3)
				return []sms.Service{svc0, svc1, svc2}
			},
			idx:         0,
			cnt:         2,
			threshold:   3,
			expectedErr: errors.New("所有短信服务商都发送失败"),
			expectedIdx: 0,
			expectedCnt: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			s := NewServiceForTimeout(tc.mock(ctrl), tc.threshold)
			s.idx = tc.idx
			s.cnt = tc.cnt
			err := s.Send(context.Background(), "test", []string{"test"}, "test")

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedIdx, s.idx)
			assert.Equal(t, tc.expectedCnt, s.cnt)
		})
	}
}
