package repository

import (
	"context"
	"database/sql"
	"errors"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *UserRepository) entityToDomain(ue dao.User) domain.User {
	return domain.User{
		Id:       ue.Id,
		Email:    ue.Email.String,
		Password: ue.Password,
		Phone:    ue.Phone.String,
	}
}

func (repo *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	ue, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(ue), nil
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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
		err = repo.cache.Set(ctx, u)
		if err != nil {
			// 打日志
		}
		return u, nil
	default:
		return domain.User{}, err
	}
}

func (repo *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	ue, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(ue), nil
}
