package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

var (
	ErrUserDuplicate          = repository.ErrUserDuplicate
	ErrInvalidEmailOrPassword = errors.New("邮箱或密码错误")
	ErrUserNotFound           = repository.ErrUserNotFound
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	// UpdateNonSensitiveInfo 更新非敏感数据
	// 你可以在这里进一步补充究竟哪些数据会被更新
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	l    logger.LoggerV1
}

func NewUserService(repo repository.UserRepository, l logger.LoggerV1) UserService {
	return &userService{
		repo: repo,
		l:    l,
	}
}

func (svc *userService) Signup(ctx context.Context, u domain.User) error {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(encrypted)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrInvalidEmailOrPassword
		}
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidEmailOrPassword
	}
	return u, nil
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	// 写法1
	// 这种是简单的写法，依赖与 Web 层保证没有敏感数据被修改
	// 也就是说，你的基本假设是前端传过来的数据就是不会修改 Email，Phone 之类的信息的。
	// return svc.repo.Update(ctx, user)

	// 写法2
	// 这种是复杂写法，依赖于 repository 中更新会忽略 0 值
	// 这个转换的意义在于，你在 service 层面上维护住了什么是敏感字段这个语义
	user.Email = ""
	user.Phone = ""
	user.Password = ""
	return svc.repo.Update(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error) {
	// 快路径，大部分请求都会进来这里
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, ErrUserNotFound) {
		// 注意 err == nil 也会来这里，返回 u
		return u, err
	}

	// 这里 phone 脱敏之后再打出来
	// zap.L().Info("手机用户未注册,注册新用户", zap.String("phone", phone))

	// svc.logger.Info("手机用户未注册,注册新用户", zap.String("phone", phone))

	svc.l.Info("手机用户未注册,注册新用户", logger.String("phone", phone))

	// 触发降级之后不执行慢路径
	// if ctx.Value("降级") == "true" {
	// 	return domain.User{}, errors.New("触发系统降级")
	// }
	// 慢路径
	// 执行注册
	err = svc.repo.Create(ctx, domain.User{Phone: phone})
	// ErrUserDuplicate 错误表明新用户已经存在，可能是并发情况下的重复创建
	if err != nil && !errors.Is(err, ErrUserDuplicate) {
		return domain.User{}, err
	}
	// 这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if !errors.Is(err, ErrUserNotFound) {
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{WechatInfo: wechatInfo})
	if err != nil && !errors.Is(err, ErrUserDuplicate) {
		return domain.User{}, err
	}
	return svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
}
