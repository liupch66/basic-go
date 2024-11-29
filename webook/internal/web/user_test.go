package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestUserHandler_LoginSms(t *testing.T) {
	now := time.Now() // 纳秒级别
	testCases := []struct {
		name string

		mock    func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBody string

		expectedResult Result
	}{
		{
			name: "手机验证码登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "15512345678", "123456").
					Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "15512345678").
					Return(domain.User{
						Id:       3,
						Email:    "123@qq.com",
						Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
						Phone:    "15512345678",
						Ctime:    now,
					}, nil)
				return userSvc, codeSvc
			},
			reqBody: `{"phone": "15512345678", "code": "123456"}`,
			expectedResult: Result{
				Code: 4,
				Msg:  "验证码验证成功",
			},
		},
		{
			name: "bind 失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBody: `{"phone": "15512345678", "code": "123456"`,
			expectedResult: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "手机格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBody: `{"phone": "155123456789", "code": "123456"}`,
			expectedResult: Result{
				Code: 4,
				Msg:  "请输入正确的手机号码",
			},
		},
		{
			name: "手机验证码过期",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "15512345678", "123456").
					Return(false, service.ErrCodeVerifyExpired)
				return nil, codeSvc
			},
			reqBody: `{"phone": "15512345678", "code": "123456"}`,
			expectedResult: Result{
				Code: 4,
				Msg:  "验证码已过期",
			},
		},
		{
			name: "手机验证码校验发生未知错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "15512345678", "123456").
					Return(false, errors.New("校验发生未知错误"))
				return nil, codeSvc
			},
			reqBody: `{"phone": "15512345678", "code": "123456"}`,
			expectedResult: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "手机验证码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "15512345678", "123456").
					Return(false, nil)
				return nil, codeSvc
			},
			reqBody: `{"phone": "15512345678", "code": "123456"}`,
			expectedResult: Result{
				Code: 4,
				Msg:  "验证码错误",
			},
		},
		{
			name: "FindOrCreateByPhone 出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "15512345678", "123456").
					Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "15512345678").
					Return(domain.User{}, errors.New("FindOrCreateByPhone 出错"))
				return userSvc, codeSvc
			},
			reqBody: `{"phone": "15512345678", "code": "123456"}`,
			expectedResult: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			server := gin.Default()
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			resp := httptest.NewRecorder()
			userHdl := NewUserHandler(tc.mock(ctrl))
			userHdl.RegisterRoutes(server)
			server.ServeHTTP(resp, req)

			var res Result
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
