package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string

		mock func(t *testing.T) *sql.DB
		ctx  context.Context
		u    User

		expectedErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				// 这边预期的是正则表达式,这个写法的意思就是,只要是 INSERT 到 users 的语句
				// sqlmock 默认是区分大小写的,写 insert into 会报错
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(sqlmock.NewResult(3, 1))
				return mockDB
			},
			u:           User{Email: sql.NullString{String: "123@qq.com", Valid: true}},
			expectedErr: nil,
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysql.MySQLError{Number: 1062})
				return mockDB
			},
			u:           User{Email: sql.NullString{String: "123@qq.com", Valid: true}},
			expectedErr: ErrUserDuplicate,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(errors.New("数据库错误"))
				return mockDB
			},
			u:           User{Email: sql.NullString{String: "123@qq.com", Valid: true}},
			expectedErr: errors.New("数据库错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
				Logger:                 logger.Default.LogMode(logger.Info),
			})
			assert.NoError(t, err)
			ud := NewUserDAO(db)
			err = ud.Insert(tc.ctx, tc.u)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
