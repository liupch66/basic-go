package job

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

// CronJob 使用别名来做一个解耦，后续万一我们要加字段，就很方便扩展
type CronJob = domain.CronJob

type Executor interface {
	Name() string
	Exec(ctx context.Context, j CronJob) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j CronJob) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j CronJob) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j CronJob) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务：%s， 你是否注册了？", j.Name)
	}
	return fn(ctx, j)
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j CronJob) error) {
	l.funcs[name] = fn
}

type Scheduler struct {
	execs     map[string]Executor
	svc       service.CronJobService
	l         logger.LoggerV1
	dbTimeout time.Duration
	// 用于控制并发数量。它提供了一个轻量级的计数信号量，用于限制资源的访问并协调多个 goroutine 之间的并发执行。
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		execs:     make(map[string]Executor),
		svc:       svc,
		l:         l,
		dbTimeout: time.Second,
		limiter:   semaphore.NewWeighted(200),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Start(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 超时了，或者被取消运行，直接退出主调度循环
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 抢占可运行的任务，数据库查询的时候，超时时间要短
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 没有抢占到，进入下一个循环，也可以睡眠一段时间
			// 也可以进一步细分错误，可以容忍就继续，否则就 return
			s.l.Error("抢占任务失败", logger.Error(err))
			continue
		}
		// 执行任务
		exec, ok := s.execs[j.Executor]
		if !ok {
			// 不支持的执行方式。比如说，这里要求的 runner 是调用 gRPC，我们就不支持
			// DEBUG 的时候最好中断，线上就继续
			s.l.Error("未找到对应的执行器", logger.String("executor: ", j.Executor))
			j.CancelFunc()
			continue
		}
		// 单独开一个 goroutine 异步执行，不要阻塞主调度循环，进入下一个循环
		go func() {
			defer func() {
				s.limiter.Release(1)
				j.CancelFunc()
			}()

			er := exec.Exec(ctx, j)
			if er != nil {
				// 也可以考虑在这里重试
				s.l.Error("调度任务执行失败", logger.Error(er), logger.Int64("jid", j.Id))
				return
			}
			// 考虑下一次调度
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				s.l.Error("设置任务的下一次执行时间失败", logger.Error(er), logger.Int64("jid", j.Id))
			}
		}()
	}
}
