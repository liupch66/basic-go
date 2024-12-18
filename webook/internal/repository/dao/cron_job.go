package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	jobStatusWaiting = iota
	jobStatusRunning
	jobStatusPaused
)

type CronJob struct {
	Id             int64  `gorm:"primaryKey,autoIncrement"`
	Name           string `gorm:"unique"`
	Cfg            string
	CronExpression string
	Executor       string
	// 标记哪些任务可以抢占，哪些已经被抢占，哪些永远不会调度之类的
	status  int
	Version int
	// 下次被调度时间
	// 查询可抢占任务条件：status = 0 AND next_time <= now
	// 建立索引，更加好的应该是 status 和 next_time 的联合索引
	NextTime int64 `gorm:"index"`
	Ctime    int64
	Utime    int64
}

type CronJobDAO interface {
	Preempt(ctx context.Context) (CronJob, error)
	UpdateUtime(ctx context.Context, jid int64) error
	Release(ctx context.Context, jid int64) error
	Stop(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

type GORMCronJobDAO struct {
	db *gorm.DB
}

func NewGORMCronJobDAO(db *gorm.DB) CronJobDAO {
	return &GORMCronJobDAO{db: db}
}

func (dao *GORMCronJobDAO) Preempt(ctx context.Context) (CronJob, error) {
	now := time.Now().UnixMilli()
	db := dao.db.WithContext(ctx)
	for {
		var j CronJob
		// 找到可抢占任务
		err := db.Where("status = ? AND next_time <= ?", jobStatusWaiting, now).First(&j).Error
		if err != nil {
			return CronJob{}, err
		}
		// 找到了可抢占任务，接下来抢占
		// 乐观锁，CAS 操作: Compare And Swap
		// 面试套路（性能优化）：曾经用了 FOR UPDATE => 性能差，还会有死锁 => 优化成了乐观锁
		res := db.Model(&CronJob{}).Where("id = ? AND version = ?", j.Id, j.Version).Updates(map[string]any{
			"status":  jobStatusRunning,
			"version": j.Version + 1,
			"utime":   now,
		})
		if res.Error != nil {
			return CronJob{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 抢占失败，继续下一轮抢占其他任务
			continue
		}
		return j, nil
	}
}

func (dao *GORMCronJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	return dao.db.WithContext(ctx).Model(&CronJob{}).Where("id = ?", jid).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMCronJobDAO) Release(ctx context.Context, jid int64) error {
	return dao.db.WithContext(ctx).Model(&CronJob{}).Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMCronJobDAO) Stop(ctx context.Context, jid int64) error {
	return dao.db.WithContext(ctx).Model(&CronJob{}).Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMCronJobDAO) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	return dao.db.WithContext(ctx).Model(&CronJob{}).Where("id = ?", jid).Updates(map[string]any{
		"next_time": nextTime.UnixMilli(),
		"utime":     time.Now().UnixMilli(),
	}).Error
}
