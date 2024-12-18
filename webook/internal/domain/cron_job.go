package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	Id             int64
	Name           string // 比如 rank
	Cfg            string
	CronExpression string
	Executor       string
	NextTime       time.Time
	CancelFunc     func() // 放弃抢占状态
}

var parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func (j *CronJob) Next(t time.Time) time.Time {
	// 这个地方 CronExpression 必须不能出错，这需要用户在注册的时候确保
	s, _ := parser.Parse(j.CronExpression)
	return s.Next(t)
}
