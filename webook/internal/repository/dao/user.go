package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("邮箱或手机冲突")
	ErrDataNotFound  = gorm.ErrRecordNotFound
)

type User struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Nickname sql.NullString
	Email    sql.NullString `gorm:"unique"`
	Password string
	// 正确处理 phone 的 NULL 值
	// 在有唯一索引的字段中，可以有多个 NULL 值，
	// 但如果字段是空字符串 ("")，则不允许有多个空字符串，数据库会将其视为相同的值，从而违反唯一索引约束
	Phone         sql.NullString `gorm:"unique"`
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString `gorm:"unique"`
	Ctime         int64
	Utime         int64
}

// TableName 自定义表名
// func (User) TableName() string {
// 	return "user"
// }

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateNonZeroFields(ctx context.Context, u User) error
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueIndexErrNo uint16 = 1062
		if mysqlErr.Number == uniqueIndexErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	// err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateNonZeroFields(ctx context.Context, u User) error {
	// 这种写法是很不清晰的，因为它依赖了 gorm 的两个默认语义
	// 会使用 ID 来作为 WHERE 条件
	// 会使用非零值来更新
	// 另外一种做法是显式指定只更新必要的字段，
	// 那么这意味着 DAO 和 service 中非敏感字段语义耦合了
	return dao.db.WithContext(ctx).Updates(&u).Error
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&u).Error
	if err != nil {
		return User{}, err
	}
	return u, nil
}
