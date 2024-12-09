package repository

import (
	"context"
	"errors"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/pkg/logger"
)

type InteractRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interact, error)
	Liked(ctx context.Context, biz string, bizId, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId, uid int64) (bool, error)
}

type CachedInteractRepository struct {
	dao   dao.InteractDAO
	cache cache.InteractCache
	l     logger.LoggerV1
}

func NewCachedInteractRepository(dao dao.InteractDAO, cache cache.InteractCache, l logger.LoggerV1) InteractRepository {
	return &CachedInteractRepository{dao: dao, cache: cache, l: l}
}

func (repo *CachedInteractRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractRepository) AddCollectionItem(ctx context.Context, biz string, bizId, cid, uid int64) error {
	err := repo.dao.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}
	// 更新缓存中的计数
	return repo.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interact, error) {
	inter, err := repo.cache.Get(ctx, biz, bizId)
	if err == nil {
		// 缓存只缓存了具体的数字，但是没有缓存自身有没有点赞的信息
		// 因为一个人反复刷，重复刷一篇文章是小概率的事情
		// 也就是说，你缓存了某个用户是否点赞的数据，命中率会很低
		return inter, nil
	}
	interEntity, err := repo.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interact{}, err
	}
	inter = domain.Interact{
		ReadCnt:    interEntity.ReadCnt,
		LikeCnt:    interEntity.LikeCnt,
		CollectCnt: interEntity.CollectCnt,
	}
	if er := repo.cache.Set(ctx, biz, bizId, inter); er != nil {
		// 可以容忍的错误
		repo.l.Error("回写缓存失败", logger.String("biz", biz),
			logger.Int64("biz_id", bizId), logger.Error(er))
	}
	return inter, nil
}

func (repo *CachedInteractRepository) Liked(ctx context.Context, biz string, bizId, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		// 吞掉 error
		return false, nil
	default:
		return false, err
	}
}

func (repo *CachedInteractRepository) Collected(ctx context.Context, biz string, bizId, uid int64) (bool, error) {
	_, err := repo.dao.GetCollectionInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		return false, nil
	default:
		return false, err
	}
}
