package job

import (
	"context"
	"sync"
	"time"

	"github.com/gotomicro/redis-lock"

	"basic-go/webook/internal/service"
	"basic-go/webook/pkg/logger"
)

type RankJob struct {
	svc        service.RankService
	timeout    time.Duration
	lockClient *rlock.Client
	key        string
	l          logger.LoggerV1
	lock       *rlock.Lock
	localLock  *sync.Mutex
}

func NewRankJob(svc service.RankService, timeout time.Duration, lockClient *rlock.Client, l logger.LoggerV1) *RankJob {
	return &RankJob{
		svc:        svc,
		timeout:    timeout,
		lockClient: lockClient,
		key:        "rlock:cron_job:rank",
		l:          l,
		localLock:  &sync.Mutex{},
	}
}

func (r *RankJob) Name() string {
	return "rank"
}

func (r *RankJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 尝试拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := r.lockClient.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: 100 * time.Millisecond,
			Max:      0,
		}, time.Second)
		// 如果没有拿到分布式锁，那就说明（大概率）有别的节点已经拿到了分布式锁
		if err != nil {
			return nil
		}
		r.lock = lock
		// 怎么保证一直拿锁
		go func() {
			// 自动续约
			er := r.lock.AutoRefresh(r.timeout/2, time.Second)
			if er != nil {
				r.l.Error("自动续约失败：", logger.Error(er))
			}
			// 下次继续抢锁
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankJob) Close() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	lock := r.lock
	r.lock = nil
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
