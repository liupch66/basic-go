package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository"
	"basic-go/webook/pkg/logger"
)

//go:generate mockgen -package=svcmocks -source=interact.go -destination=mocks/interact_mock.go InteractService
type InteractService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interact, error)
	// GetByIds 这里本来返回 []domain.Interact，返回 map 是方便查找对应文章 id 的点赞数据
	GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interact, error)
}

type interactService struct {
	repo repository.InteractRepository
	l    logger.LoggerV1
}

func NewInteractService(repo repository.InteractRepository, l logger.LoggerV1) InteractService {
	return &interactService{repo: repo, l: l}
}

func (svc *interactService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizId)
}

func (svc *interactService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return svc.repo.IncrLike(ctx, biz, bizId, uid)
}

func (svc *interactService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return svc.repo.DecrLike(ctx, biz, bizId, uid)
}

func (svc *interactService) Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	return svc.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (svc *interactService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interact, error) {
	// 你也可以考虑将分发的逻辑也下沉到 repository 里面
	inter, err := svc.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interact{}, err
	}
	if uid > 0 {
		var (
			eg        errgroup.Group
			liked     bool
			collected bool
		)

		eg.Go(func() error {
			var er error
			liked, er = svc.repo.Liked(ctx, biz, bizId, uid)
			return er
		})
		eg.Go(func() error {
			var er error
			collected, er = svc.repo.Collected(ctx, biz, bizId, uid)
			return er
		})
		err = eg.Wait()
		if err != nil {
			svc.l.Error("查询用户是否点赞的信息失败", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Int64("uid", uid), logger.Error(err))
		}
		inter.Liked = liked
		inter.Collected = collected
	}
	return inter, nil
}

func (svc *interactService) GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interact, error) {
	// TODO implement me
	panic("implement me")
}
