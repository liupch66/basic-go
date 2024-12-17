package ioc

import (
	"time"

	"github.com/robfig/cron/v3"

	"basic-go/webook/internal/job"
	"basic-go/webook/internal/service"
	"basic-go/webook/pkg/logger"
)

func InitRankJob(svc service.RankService) *job.RankJob {
	return job.NewRankJob(svc, 30*time.Second)
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
