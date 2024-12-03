package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"basic-go/webook/internal/integration/startup"
	"basic-go/webook/internal/web"
	"basic-go/webook/ioc"
)

func TestUserHandler_SendSmsLoginCode(t *testing.T) {
	server := startup.InitWebServer()
	cmd := ioc.InitRedis()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		reqBody string

		expectedCode int
		expectedRes  web.Result
	}{
		{
			name:   "发送成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				val, err := cmd.GetDel(context.Background(), "phone_code:login:15512345678").Result()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody:      `{"phone": "15512345678"}`,
			expectedCode: http.StatusOK,
			expectedRes:  web.Result{Code: 4, Msg: "发送成功"},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				cmd.Set(ctx, "phone_code:login:15512345678", "123456", time.Minute*9+time.Second*30)
			},
			after: func(t *testing.T) {
				val, err := cmd.GetDel(context.Background(), "phone_code:login:15512345678").Result()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody:      `{"phone": "15512345678"}`,
			expectedCode: http.StatusOK,
			expectedRes:  web.Result{Code: 4, Msg: "发送验证码太频繁"},
		},
		{
			name:         "手机号码格式错误",
			before:       func(t *testing.T) {},
			after:        func(t *testing.T) {},
			reqBody:      `{"phone": "155123456789"}`,
			expectedCode: http.StatusOK,
			expectedRes:  web.Result{Code: 4, Msg: "请输入正确的手机号码"},
		},
		{
			name:         "bind 失败",
			before:       func(t *testing.T) {},
			after:        func(t *testing.T) {},
			reqBody:      `{"phone": "15512345678"`,
			expectedCode: http.StatusBadRequest,
			expectedRes:  web.Result{Code: 5, Msg: "系统错误"},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 设置没有过期时间
				cmd.Set(context.Background(), "phone_code:login:15512345678", "123456", 0)
			},
			after: func(t *testing.T) {
				val, err := cmd.GetDel(context.Background(), "phone_code:login:15512345678").Result()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody:      `{"phone": "15512345678"}`,
			expectedCode: http.StatusOK,
			expectedRes:  web.Result{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			var res web.Result
			// json.NewDecoder(resp.Body).Decode(&res) 是流式解码，它不会一次性加载整个响应体到内存，而是逐步读取和解码。
			// 这使得它在处理大文件或大量数据时更高效，尤其是当响应体很大的时候。
			err = json.NewDecoder(resp.Body).Decode(&res)
			// json.Unmarshal(resp.Body.Bytes(), &res) 会先读取整个 resp.Body 到内存中，再解码，适用于响应体较小的场景。
			// 对于大文件，这可能导致较高的内存消耗。
			// err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedRes, res)

			tc.after(t)
		})
	}
}
