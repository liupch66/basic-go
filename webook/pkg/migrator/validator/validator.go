package validator

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/migrator"
	events2 "github.com/liupch66/basic-go/webook/pkg/migrator/events"
)

// Validator T 必须实现了 Entity 接口
// 源表 src 和目标表 dst 不会变，但是 base 和 target 会变，direction 也会变
type Validator[T migrator.Entity] struct {
	// 校验，以 XX 为准
	base *gorm.DB
	// 校验谁的数据
	target *gorm.DB

	// 这边需要告知，是以 SRC 为准，还是以 DST 为准，修复数据需要知道
	direction string

	batchSize int

	l        logger.LoggerV1
	producer events2.Producer

	highLoad *atomic.Bool

	// 根据 utime 和 sleepInterval 的组合，就可以同时支持全量校验和增量校验。
	// utime = 0 并且 sleepInterval <= 0：那么就是全量校验，并且在数据校验完毕之后，就直接退出。
	// utime = 0 并且 sleepInterval > 0：那么就是全量校验，并且在全量校验之后，还会继续增量校验。
	// utime 是近期某个时间点（utime > 0），并且 sleepInterval > 0：就是校验近期修改的数据，并且后续保持增量校验。
	utime         int64
	sleepInterval time.Duration

	// 全量校验 id 合适，增量校验 utime 合适
	order string
}

// option 模式
// 写法一优缺点：
// 因为 Option 函数类型和 Validator 类型都使用了泛型，你可以为不同类型的 Validator 创建配置选项，而不局限于
// 某个具体类型（如 migrator.Entity）。如果你有多个不同类型的 Validator，可以使用 Option 来为它们分别配置。
// 对于较简单的应用场景，如果只处理 Validator[migrator.Entity] 类型，这种写法显得有些过度设计。

type Option[T migrator.Entity] func(*Validator[T])

func WithUtime[T migrator.Entity](utime int64) Option[T] {
	return func(v *Validator[T]) {
		v.utime = utime
	}
}

func WithSleepInterval[T migrator.Entity](sleepInterval time.Duration) Option[T] {
	return func(v *Validator[T]) {
		v.sleepInterval = sleepInterval
	}
}

func WithOrder[T migrator.Entity](order string) Option[T] {
	return func(v *Validator[T]) {
		v.order = order
	}
}

// 写法二优缺点：
// 简单易懂，容易实现。当确定 Option 只需要针对 Validator[migrator.Entity] 类型时，这种写法可以满足需求。
// 泛型 T 没有得到充分利用，缺少了灵活性。如果将来需要针对不同类型的 Validator 使用 Option，这种方式就会受到限制。

// type Option func(*Validator[migrator.Entity])
//
// func WithUtime(utime int64) Option {
// 	return func(v *Validator[migrator.Entity]) {
// 		v.utime = utime
// 	}
// }
//
// func WithSleepInterval(sleepInterval time.Duration) Option {
// 	return func(v *Validator[migrator.Entity]) {
// 		v.sleepInterval = sleepInterval
// 	}
// }
//
// func WithOrder(order string) Option {
// 	return func(v *Validator[migrator.Entity]) {
// 		v.order = order
// 	}
// }

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string,
	l logger.LoggerV1, producer events2.Producer, opts ...Option[T]) *Validator[T] {
	highLoad := atomic.NewBool(false)
	go func() {
		// 查询数据库性能，看是否高负载
		// 性能瓶颈一般在数据库，也可以结合 CPU，内存负载动态判定
	}()
	v := &Validator[T]{
		base:      base,
		target:    target,
		direction: direction,
		batchSize: 100,
		l:         l,
		producer:  producer,
		highLoad:  highLoad,
		// 默认是全量校验，并且数据没了就结束
		utime:         0,
		sleepInterval: 0,
		// 默认是 id，下面增量校验的时候可能需要 utime
		order: "id",
	}
	for _, opt := range opts {
		opt(v)
		// opt((*Validator[migrator.Entity])(v))
	}
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		v.baseToTarget(ctx)
		return nil
	})
	eg.Go(func() error {
		v.targetToBase(ctx)
		return nil
	})
	return eg.Wait()
}

// baseToTarget 执行 base 到 target 的验证，找出 dst 中不一致和没有的数据
func (v *Validator[T]) baseToTarget(ctx context.Context) {
	offset := 0
	for {
		if v.highLoad.Load() {
			// 挂起
		}
		// 先查询源表
		var src T
		// 有 offset 设置一个 order 保证 offset 的稳定性
		// 最好不要取等号。比如增量校验时设置 v.utime = 12:00（现在），手速够快，
		// 第一次进来刚好还没更改， utime = 12:00，第二次修改了进来 utime > 12:00，
		// 因为是查询结果按 utime 排序，导致重复检验事小，主要可能会导致很多漏检，尤其是下面反过来分批次查询的时候，
		// 这里是一个个查。当然这种情况不太可能，但是还是会有类似的问题，就是说 v.utime 不好设置
		err := v.base.WithContext(ctx).Where("utime > ?", v.utime).Offset(offset).
			Order(v.order).First(&src).Error
		switch {
		// 里面 switch 不加这个分支也行，下个循环还是会到这里
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			return
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 已经没有数据了，全量校验结束了，直接 return
			// return
			// 但是要同时支持全量校验和增量校验，这里就不能直接返回，因为业务会持续产生新数据，所以增量校验会一直运行
			// 在这里要考虑，用户有些情况下希望退出，有些情况下希望继续
			// 当用户希望继续时，防止下一个 for 循环进入这里还是没数据，所以要 sleep 一下
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		case err == nil:
			// 再查询目标表
			var dst T
			err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
			switch {
			case err == nil:
				// 查询到了，怎么比较？不能直接比较 src == dst
				// // 1. 直接利用反射来比较，原则上可以
				// if !reflect.DeepEqual(src, dst) {
				// 	v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
				// }

				// 2. 自己实现 Entity 的比较逻辑
				if !src.CompareTo(dst) {
					v.notify(src.ID(), events2.InconsistentEventTypeNotEqual)
				}

				// 3. 判断有没有实现自己的比较逻辑
				// var srcAny any = src
				// if s, ok := srcAny.(interface {
				// 	CompareTo(dst migrator.Entity) bool
				// }); ok {
				// 	if !s.CompareTo(dst) {
				// 		v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
				// 	}
				// } else {
				// 	if !reflect.DeepEqual(src, dst) {
				// 		v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
				// 	}
				// }

			case errors.Is(err, gorm.ErrRecordNotFound):
				v.notify(src.ID(), events2.InconsistentEventTargetMissing)
			default:
				v.l.Error("src => dst 查询目标表失败", logger.Error(err))
			}
		default:
			// 查询有错，offset 加下轮就跳过这条，不跳很可能一直在这循环（比如数据库有问题，
			// 但是数据库有问题跳过好像也没啥用？这里还得考虑），两者相比还是跳过
			v.l.Error("src => dst 查询源表失败", logger.Error(err))
		}
		offset++
	}
}

// targetToBase 反过来，执行 target 到 base 的验证，找出 dst 中多余的数据
// 出现情况：在同步数据到 dstDB 之后，srcDB 中的数据就被删除了，那么 dstDB 就会多了一些数据。
// 注意，这种删除必须是硬删除，软删除本质上是一个 UPDATE，所以不会有这个问题

// 理论上来说，可以利用 count 来加速这个过程，这一步大多数情况下效果很好，尤其是那些软删除的
// 举个例子，假如初始化目标表的数据是昨天的 23:59:59 导出来的
// 那么可以 COUNT(*) WHERE ctime < 今天的零点，count 如果相等，就说明没删除。
// 如果 count 不一致，还得一条条查出来

// todo: 这里为什么不先把 base 中的 id 统计出来，再从 target 中找没有的呢？
// 应该是因为这样的话就是一直都是全量校验，而我们是想同时实现全量校验（比较昂贵）和增量校验
// 整个全量校验和修复可以看做是两个步骤：校验，如果发现不一致，则修复。因此从形态上来说，有以下几种比较典型的做法：
// • 校验如果发现数据不一致，那么立刻修复。这些都是同一个 goroutine 来执行的。
// • 校验如果发现数据不一致，那么立刻交给另外一个 goroutine 去修复。可以引入 channel，也可以不引入 channel。
// • 校验如果发现数据不一致，那么发送消息到消息队列中，消费者消费了再去修复数据。（我的方案）

func (v *Validator[T]) targetToBase(ctx context.Context) {
	offset := 0
	for {
		var dstTs []T
		// 这里注意一种情况：数据 A 在 base 和 target 中， utime = 昨天
		// 这里增量校验今天新产生的数据，utime = 今天 0 点，base 今天又刚好删了数据 A，
		// 但是 Where("utime > ?", v.utime) 检查不到，就会导致 target 多了一个数据
		// 解决办法：这个时候 utime 可能要稍微调小一点，但是也不好调，调多小呢，硬删除就是比较麻烦
		// 只能用 canal 来监听 binlog，记录了所有改变数据库状态的 SQL 语句（如 INSERT、UPDATE、DELETE 等）
		// 不过只要 base 在删除 A 之前进行过修改，utime = 今天，这个就没有影响
		// 注意： utime 上面如果没有一个独立的索引，那么在查询的时候就会特别慢。
		// 增量校验的时候，这里 order by utime 好一点
		err := v.target.WithContext(ctx).Model(new(T)).Where("utime > ?", v.utime).
			Order(v.order).Offset(offset).Limit(v.batchSize).Find(&dstTs).Error

		switch {
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			return
		// 坑：还是要做判断，find 这里不会返回 gorm.ErrRecordNotFound，只是会返回空切片，first, last, take 会返回
		// 正常来说，Find(&dstTs) 不会有这个 gorm.ErrRecordNotFound，保险起见
		case errors.Is(err, gorm.ErrRecordNotFound) || len(dstTs) == 0:
			// 全量校验没有数据就返回
			// return
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			// 这里 offset 不能加
			continue
		case err == nil:
			dstIds := slice.Map(dstTs, func(idx int, dst T) int64 {
				return dst.ID()
			})
			var srcTs []T
			err = v.base.WithContext(ctx).Model(new(T)).Where("id IN ?", dstIds).Find(&srcTs).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				v.notifyBaseMissing(dstTs)
			case err == nil:
				// 计算差集
				diff := slice.DiffSetFunc(dstTs, srcTs, func(dst, src T) bool {
					return dst.ID() == src.ID()
				})
				v.notifyBaseMissing(diff)
			default:
				v.l.Error("dst => src 查询源表失败", logger.Error(err))
			}

		default:
			v.l.Error("dst => src 查询目标表失败", logger.Error(err))
		}
		if len(dstTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			// 这里 offset 要加 len(dstTs)
		}
		offset += len(dstTs)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	// 这里我们要单独控制超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events2.InconsistentEvent{
		Id:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		// 可以重试，但是重试也会失败，记日志，告警，手动去修
		// 也可以直接忽略，下一轮修复和校验又会找出来
		v.l.Error("发送消息失败", logger.Error(err))
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, t := range ts {
		v.notify(t.ID(), events2.InconsistentEventBaseMissing)
	}
}

// 通用写法，摆脱对 T 的依赖
// func (v *Validator[T]) interactFromBase(ctx context.Context, offset int) (T, error) {
//	rows, err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
//		Offset(offset).Order("utime ASC, id ASC").Rows()
//	cols, err := rows.Columns()
//	// 所有列的值
//	vals := make([]any, len(cols))
//	rows.Scan(vals...)
//	return vals
// }
