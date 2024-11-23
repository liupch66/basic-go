package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
	svcmocks "basic-go/webook/internal/service/mocks"
)

// require.NoError(t, err)：如果 err 为 nil，测试继续。如果 err 不为 nil，会立即导致测试失败并停止后续代码执行。
// 适用于那些后续测试无法继续执行的情况（例如依赖于某个操作是否成功）。
// assert.NoError(t, err)：如果 err 为 nil，测试继续。如果 err 不为 nil，会记录失败，但测试函数会继续执行。
// 适用于收集多个错误或希望确保执行完所有测试的场景。

func TestUserHandler_Signup(t *testing.T) {
	testCases := []struct {
		name    string
		reqBody string
		mock    func(ctrl *gomock.Controller) service.UserService

		expectedCode int
		expectedBody string
	}{
		{
			name:    "注册成功",
			reqBody: `{"email": "123@qq.com", "password": "hello#world123", "confirmPassword": "hello#world123"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).
					Return(nil)
				return userSvc
			},
			expectedCode: http.StatusOK,
			expectedBody: "注册成功",
		},
		{
			name:    "bind 失败",
			reqBody: `{"email": "123@qq.com", "password": "hello#world123", "confirmPassword": "hello#world123"`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				// bind 失败没有走到 signup 的调用, 也可以偷懒不 mock userSvc
				return nil
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:    "邮箱格式错误",
			reqBody: `{"email": "123qq.com", "password": "hello#world123", "confirmPassword": "hello#world123"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {

				return nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "邮箱格式错误",
		},
		{
			name:    "密码格式错误",
			reqBody: `{"email": "123@qq.com", "password": "helloworld123", "confirmPassword": "helloworld123"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "密码必须大于 8 位，并且包含数字，字母和特殊符号",
		},
		{
			name:    "两次输入的密码不一致",
			reqBody: `{"email": "123@qq.com", "password": "hello#world123", "confirmPassword": "hello#world12"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "两次输入的密码不一致",
		},
		{
			name:    "注册邮箱重复",
			reqBody: `{"email": "123@qq.com", "password": "hello#world123", "confirmPassword": "hello#world123"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).
					Return(service.ErrUserDuplicate)
				return userSvc
			},
			expectedCode: http.StatusOK,
			expectedBody: "邮箱重复，请换一个邮箱",
		},
		{
			name:    "系统错误",
			reqBody: `{"email": "123@qq.com", "password": "hello#world123", "confirmPassword": "hello#world123"}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).
					Return(errors.New("nothing"))
				return userSvc
			},
			expectedCode: http.StatusOK,
			expectedBody: "系统错误",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			resp := httptest.NewRecorder()
			// 这里用不上 codeSvc, 可以偷懒,不 mock 出来
			userHdl := NewUserHandler(tc.mock(ctrl), nil)
			userHdl.RegisterRoutes(server)
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedBody, resp.Body.String())
		})
	}
}
