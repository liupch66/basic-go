package ioc

import (
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"

	"github.com/liupch66/basic-go/webook/internal/job"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

func InitRankJob(svc service.RankService, lockClient *rlock.Client, l logger.LoggerV1) *job.RankJob {
	return job.NewRankJob(svc, 30*time.Second, lockClient, l)
}

func InitJobs(l logger.LoggerV1, rankJob *job.RankJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	b := job.NewCronJobBuilder(l)
	// 秒 分 时 日 月 星期 （年）
	_, err := res.AddJob("0 */3 * * * ?", b.Build(rankJob))
	if err != nil {
		panic(err)
	}
	return res
}
