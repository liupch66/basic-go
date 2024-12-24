package validator

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/migrator"
	"github.com/liupch66/basic-go/webook/migrator/events"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

// Validator T 必须实现了 Entity 接口
type Validator[T migrator.Entity] struct {
	// 校验，以 XX 为准
	base *gorm.DB
	// 校验谁的数据
	target *gorm.DB

	// 这边需要告知，是以 SRC 为准，还是以 DST 为准，修复数据需要知道
	direction string

	batchSize int

	l        logger.LoggerV1
	producer events.Producer

	highLoad *atomic.Bool
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string,
	l logger.LoggerV1, producer events.Producer) *Validator[T] {
	highLoad := atomic.NewBool(false)
	go func() {
		// 性能瓶颈一般在数据库，也可以结合 CPU，内存负载动态判定
	}()
	return &Validator[T]{
		base:      base,
		target:    target,
		direction: direction,
		batchSize: 100,
		l:         l,
		producer:  producer,
		highLoad:  highLoad,
	}
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
	offset := -1
	for {
		// 开始就自增更简洁，入口唯一，出口太多：continue, return
		offset++

		// 先查询源表
		var src T
		err := v.base.WithContext(ctx).Offset(offset).Order("id").First(&src).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 已经没有数据了
			return
		}
		if err != nil {
			v.l.Error("src => dst 查询源表失败", logger.Error(err))
			continue
		}

		// 再查询目标表
		var dst T
		err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.notify(src.ID(), events.InconsistentEventTargetMissing)
		}
		if err != nil {
			v.l.Error("src => dst 查询目标表失败", logger.Error(err))
			continue
		}

		// 查询到了，怎么比较？不能直接比较 src == dst
		// // 1. 直接利用反射来比较
		// if !reflect.DeepEqual(src, dst) {
		// 	v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
		// }

		// 2. 自己实现 Entity 的比较逻辑
		if !src.CompareTo(dst) {
			v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
		}

		// // 3. 判断有没有实现自己的比较逻辑
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
	}

}

// targetToBase 反过来，执行 target 到 base 的验证，找出 dst 中多余的数据
func (v *Validator[T]) targetToBase(ctx context.Context) {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var dstTs []T
		err := v.target.WithContext(ctx).Model(new(T)).Select("id").Offset(offset).
			Limit(v.batchSize).Find(&dstTs).Error
		// 下面做了判断，这里应该不用
		// if len(dstTs) == 0 {
		// 	return
		// }
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return
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
			continue
		}

		if len(dstTs) < v.batchSize {
			return
		}
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	// 这里我们要单独控制超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
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
		v.notify(t.ID(), events.InconsistentEventBaseMissing)
	}
}
