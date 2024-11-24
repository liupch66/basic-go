package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/repository/cache/redismocks"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name string

		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		expectedErr error
	}{
		{
			name: "设置验证码成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				// 注意传入 int64, 不然报错: redis: unexpected type=int for Int
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:15512345678"}, "123").Return(res)
				return cmd
			},
			ctx:         context.Background(),
			biz:         "login",
			phone:       "15512345678",
			code:        "123",
			expectedErr: nil,
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:15512345678"}, "123").Return(res)
				return cmd
			},
			ctx:         context.Background(),
			biz:         "login",
			phone:       "15512345678",
			code:        "123",
			expectedErr: ErrCodeSendTooMany,
		},
		{
			name: "redis 执行脚本出错",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("redis 执行脚本出错"))
				cmd.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:15512345678"}, "123").Return(res)
				return cmd
			},
			ctx:         context.Background(),
			biz:         "login",
			phone:       "15512345678",
			code:        "123",
			expectedErr: errors.New("redis 执行脚本出错"),
		},
		{
			name: "未知错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(-2))
				cmd.EXPECT().Eval(context.Background(), luaSetCode, []string{"phone_code:login:15512345678"}, "123").Return(res)
				return cmd
			},
			ctx:         context.Background(),
			biz:         "login",
			phone:       "15512345678",
			code:        "123",
			expectedErr: ErrUnknownForCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			codeCache := NewCodeCache(tc.mock(ctrl))
			err := codeCache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
