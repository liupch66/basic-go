package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository"
	repomocks "github.com/liupch66/basic-go/webook/internal/repository/mocks"
)

func TestEncrypt(t *testing.T) {
	passwd := []byte("hello#world123")
	encrypted, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, passwd)
	assert.NoError(t, err)
}

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name string

		mock     func(ctrl *gomock.Controller) repository.UserRepository
		email    string
		password string

		expectedUser domain.User
		expectedErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{
						Id:       1,
						Email:    "123@qq.com",
						Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
						Phone:    "15512345678",
					}, nil)
				return userRepo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			expectedUser: domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
				Phone:    "15512345678",
			},
			expectedErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, ErrUserNotFound)
				return userRepo
			},
			email:        "123@qq.com",
			password:     "hello#world123",
			expectedUser: domain.User{},
			expectedErr:  ErrInvalidEmailOrPassword,
		},
		{
			name: "DB 错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("DB 出错了"))
				return userRepo
			},
			email:        "123@qq.com",
			password:     "hello#world123",
			expectedUser: domain.User{},
			expectedErr:  errors.New("DB 出错了"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{
						Id:       1,
						Email:    "123@qq.com",
						Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
						Phone:    "15512345678",
					}, nil)
				return userRepo
			},
			email:        "123@qq.com",
			password:     "hello#world1234",
			expectedUser: domain.User{},
			expectedErr:  ErrInvalidEmailOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSvc := NewUserService(tc.mock(ctrl), nil)
			u, err := userSvc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.expectedUser, u)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
