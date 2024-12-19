package ioc

import (
	"context"
	"time"

	"github.com/liupch66/basic-go/webook/internal/job"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

func InitLocalFuncExecutor(svc service.RankService) *job.LocalFuncExecutor {
	executor := job.NewLocalFuncExecutor()
	executor.RegisterFunc("rank", func(ctx context.Context, j job.CronJob) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		return svc.TopN(ctx)
	})
	return executor
}

func InitScheduler(svc service.CronJobService, l logger.LoggerV1, executor *job.LocalFuncExecutor) *job.Scheduler {
	s := job.NewScheduler(svc, l)
	// 要在数据库里面插入一条 rank job 的记录，通过管理任务接口来插入
	s.RegisterExecutor(executor)
	return s
}
