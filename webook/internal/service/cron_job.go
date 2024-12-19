package service

import (
	"context"
	"time"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	ResetNextTime(ctx context.Context, j domain.CronJob) error
	// 两种释放方法：返回一个释放的方法，然后调用者去调；定义一个释放方法
	// PreemptV1(ctx context.Context) (domain.CronJob, func() error,  error)
	// Release(ctx context.Context, id int64) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{
		repo:            repo,
		refreshInterval: 10 * time.Second,
		l:               l,
	}
}

func (svc *cronJobService) Preempt(ctx context.Context) (domain.CronJob, error) {
	// 先去抢占一个 cronjob
	j, err := svc.repo.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}
	// 续约
	tc := time.NewTicker(svc.refreshInterval)
	go func() {
		for range tc.C {
			svc.refresh(j.Id)
		}
	}()
	// 释放
	j.CancelFunc = func() {
		tc.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := svc.repo.Release(ctx, j.Id)
		if err != nil {
			svc.l.Error("释放任务失败", logger.Error(err), logger.Int64("jid", j.Id))
		}
	}
	return j, err
}

func (svc *cronJobService) refresh(jid int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := svc.repo.UpdateUtime(ctx, jid)
	if err != nil {
		svc.l.Error("任务续约失败", logger.Error(err), logger.Int64("jid", jid))
	}
}

func (svc *cronJobService) ResetNextTime(ctx context.Context, j domain.CronJob) error {
	nextTime := j.Next(time.Now())
	if nextTime.IsZero() {
		return svc.repo.Stop(ctx, j.Id)
	}
	return svc.repo.UpdateNextTime(ctx, j.Id, nextTime)
}
