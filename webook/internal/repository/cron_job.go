package repository

import (
	"context"
	"time"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	UpdateUtime(ctx context.Context, jid int64) error
	Release(ctx context.Context, jid int64) error
	Stop(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

type PreemptCronJobRepository struct {
	dao dao.CronJobDAO
}

func NewPreemptCronJobRepository(dao dao.CronJobDAO) CronJobRepository {
	return &PreemptCronJobRepository{dao: dao}
}

func (repo *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.CronJob, error) {
	j, err := repo.dao.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}
	return domain.CronJob{
		Id:  j.Id,
		Cfg: j.Cfg,
	}, nil
}

func (repo *PreemptCronJobRepository) UpdateUtime(ctx context.Context, jid int64) error {
	return repo.dao.UpdateUtime(ctx, jid)
}

func (repo *PreemptCronJobRepository) Release(ctx context.Context, jid int64) error {
	return repo.dao.Release(ctx, jid)
}

func (repo *PreemptCronJobRepository) Stop(ctx context.Context, jid int64) error {
	return repo.dao.Stop(ctx, jid)
}

func (repo *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	return repo.dao.UpdateNextTime(ctx, jid, nextTime)
}
