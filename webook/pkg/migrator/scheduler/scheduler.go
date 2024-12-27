package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/pkg/ginx"
	"github.com/liupch66/basic-go/webook/pkg/gormx/connpool"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/migrator"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events"
	"github.com/liupch66/basic-go/webook/pkg/migrator/validator"
)

// Scheduler 用来统一管理整个迁移过程
// 它不是必须的，这是为了方便用户操作（和自己理解）而引入的。
type Scheduler[T migrator.Entity] struct {
	mu         sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	// 如果要允许多个全量校验同时运行
	fulls map[string]func()
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB, l logger.LoggerV1, pattern string,
	producer events.Producer, pool *connpool.DoubleWritePool) *Scheduler[T] {
	return &Scheduler[T]{
		src:     src,
		dst:     dst,
		l:       l,
		pattern: pattern,
		// 初始化的时候，不用干啥
		cancelFull: func() {},
		cancelIncr: func() {},
		producer:   producer,
		pool:       pool,
	}
}

func (s *Scheduler[T]) UpdatePattern(pattern string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pattern = pattern
	s.pool.UpdatePattern(pattern)
}

// RegisterRoutes 这一个也不是必须的，可以考虑利用配置中心，监听配置中心的变化
// 把全量校验，增量校验做成分布式任务，利用分布式任务调度平台来调度
// 在这里，有多少张表，就初始化多少个 scheduler，所以传入 *gin.RouterGroup 好一点，传 gin.Engine 没法区分各个路由
func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	// 将这个暴露为 HTTP 接口，可以配上对应的 UI
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.DstFirst))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/start", ginx.WrapReq[StartIncrRequest](s.StartIncrementValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))
}

// ---- 下面是四个阶段 ---- //

// SrcOnly 只读写源表
func (s *Scheduler[T]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{Msg: "successfully switched to src_only"}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{Msg: "successfully switched to src_first"}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{Msg: "successfully switched to dst_first"}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{Msg: "successfully switched to dst_only"}, nil
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 构造全量校验 validator，option 部分都是默认值
	var v *validator.Validator[T]
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		// utime = 0 并且 sleepInterval <= 0：那么就是全量校验，并且在数据校验完毕之后，就直接退出。
		v = validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer)
		// utime = 0 并且 sleepInterval > 0：那么就是全量校验，并且在全量校验之后，还会继续增量校验。
		// v = validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer,
		// 	validator.WithSleepInterval[T](time.Second))
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		v = validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer)
	}

	// 开启全量校验
	go func() {
		// 取消上一次的全量校验
		s.cancelFull()
		var ctx context.Context
		ctx, s.cancelFull = context.WithCancel(context.Background())
		err := v.Validate(ctx)
		s.l.Warn("退出全量校验", logger.Error(err))
	}()
	return ginx.Result{Msg: "启动全量校验"}, nil
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelFull()
	return ginx.Result{Msg: "停止全量校验"}, nil
}

type StartIncrRequest struct {
	Utime int64 `json:"utime"`
	// 毫秒数，json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}

// StartIncrementValidation 增量校验
func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrRequest) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 利用 option 模式构造增量校验 validator
	opts := []validator.Option[T]{
		validator.WithUtime[T](req.Utime),
		validator.WithSleepInterval[T](time.Duration(req.Interval) * time.Millisecond),
		validator.WithOrder[T]("utime"),
	}
	var v *validator.Validator[T]
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		v = validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer, opts...)
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		v = validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer, opts...)
	}

	// 开启增量校验
	go func() {
		// 取消上一次的增量校验
		s.cancelIncr()
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	return ginx.Result{Msg: "启动增量校验"}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelIncr()
	return ginx.Result{Msg: "停止增量校验"}, nil
}
