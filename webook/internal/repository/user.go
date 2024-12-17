package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

//go:generate mockgen -package=repomocks -source=user.go -destination=mocks/user_mock.go UserRepository
type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CachedUserRepository) entityToDomain(ue dao.User) domain.User {
	return domain.User{
		Id:       ue.Id,
		Email:    ue.Email.String,
		Password: ue.Password,
		Nickname: ue.Nickname.String,
		Phone:    ue.Phone.String,
		WechatInfo: domain.WechatInfo{
			OpenId:  ue.WechatOpenId.String,
			UnionId: ue.WechatUnionId.String,
		},
		Ctime: time.UnixMilli(ue.Ctime),
	}
}

func (repo *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Nickname: sql.NullString{
			String: u.Nickname,
			Valid:  u.Nickname != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	ue, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(ue), nil
}

func (repo *CachedUserRepository) Update(ctx context.Context, u domain.User) error {
	err := repo.dao.UpdateNonZeroFields(ctx, repo.domainToEntity(u))
	if err != nil {
		return err
	}
	return repo.cache.Delete(ctx, u.Id)
}

func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 这里注意处理方式
	u, err := repo.cache.Get(ctx, id)
	switch {
	case err == nil:
		return u, nil
	case errors.Is(err, cache.ErrKeyNotExist):
		// ue: user entity
		ue, err := repo.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		u = repo.entityToDomain(ue)
		// err = repo.cache.Set(ctx, u)
		// if err != nil {
		// 	// 打日志
		// }
		_ = repo.cache.Set(ctx, u)
		return u, nil
	default:
		return domain.User{}, err
	}
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	ue, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(ue), nil
}

func (repo *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	ue, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(ue), nil
}
