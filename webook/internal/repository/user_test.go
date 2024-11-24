package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	cachemocks "basic-go/webook/internal/repository/cache/mocks"
	"basic-go/webook/internal/repository/dao"
	daomocks "basic-go/webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	// 注意如果用 time.Now() 返回的是纳秒, time.UnixMilli(now) 会去掉毫秒以外的部分，确保返回的时间只包含毫秒级别的精度
	now := time.Now().UnixMilli()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx  context.Context
		id   int64

		expectedUser domain.User
		expectedErr  error
	}{
		{
			name: "缓存未命中,数据库查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindById(context.Background(), int64(1)).Return(dao.User{
					Id:       1,
					Email:    sql.NullString{String: "123@qq.com", Valid: true},
					Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
					Phone:    sql.NullString{String: "15512345678", Valid: true},
					Ctime:    now,
					Utime:    now,
				}, nil)
				uc.EXPECT().Set(context.Background(), domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
					Phone:    "15512345678",
					Ctime:    time.UnixMilli(now),
				}).Return(nil)
				return ud, uc
			},
			ctx: context.Background(),
			id:  1,
			expectedUser: domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
				Phone:    "15512345678",
				Ctime:    time.UnixMilli(now),
			},
			expectedErr: nil,
		},
		{
			name: "缓存直接命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
					Phone:    "15512345678",
					Ctime:    time.UnixMilli(now),
				}, nil)
				return nil, uc
			},
			ctx: context.Background(),
			id:  1,
			expectedUser: domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "$2a$10$lTkuFqE9cuX.DvR69q00OucWAtQjpzJwHGsd9mcQRvvvku7HRtQ1G",
				Phone:    "15512345678",
				Ctime:    time.UnixMilli(now),
			},
			expectedErr: nil,
		},
		{
			name: "查询缓存出错",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{}, errors.New("查询缓存出错"))
				return nil, uc
			},
			ctx:          context.Background(),
			id:           1,
			expectedUser: domain.User{},
			expectedErr:  errors.New("查询缓存出错"),
		},
		{
			name: "缓存未命中,数据库查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindById(context.Background(), int64(1)).Return(dao.User{}, errors.New("查询数据库失败"))
				return ud, uc
			},
			ctx:          context.Background(),
			id:           1,
			expectedUser: domain.User{},
			expectedErr:  errors.New("查询数据库失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// defer ctrl.Finish()

			userRepo := NewUserRepository(tc.mock(ctrl))
			u, err := userRepo.FindById(tc.ctx, tc.id)

			assert.Equal(t, tc.expectedUser, u)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
