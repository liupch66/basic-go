package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/service/sms"
	smsmocks "basic-go/webook/internal/service/sms/mocks"
	"basic-go/webook/pkg/ratelimit"
	limitmocks "basic-go/webook/pkg/ratelimit/mocks"
)

func TestService_Send(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter)

		expectedErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), key).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, limiter
			},
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), key).Return(true, nil)
				return nil, limiter
			},
			expectedErr: errLimited,
		},
		{
			name: "限流器异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), key).Return(false, errors.New("限流器异常"))
				return nil, limiter
			},
			expectedErr: fmt.Errorf("短信服务判断限流出现错误: %w", errors.New("限流器异常")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			smsSvc := NewService(tc.mock(ctrl))
			err := smsSvc.Send(context.Background(), "test tplId", []string{"test params"}, "test number")
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
