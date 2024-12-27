package fixer

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/liupch66/basic-go/webook/pkg/migrator"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*Fixer[T], error) {
	var dst T
	rows, err := target.Model(&dst).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return &Fixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (f *Fixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var src T
	err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&src).Error
	switch {
	case err == nil:
		// 修复数据的时候，可以考虑增加 WHERE base.Utime >= target.Utime 或者 version 之类的条件
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&src).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return f.target.WithContext(ctx).Where("id = ?", src.ID()).Delete(new(T)).Error
		// 简写
		// return f.target.WithContext(ctx).Delete(new(T), src.ID()).Error
	default:
		return err
	}
}

// FixV1 看上去会更加符合直觉，但是有点多余的代码
func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTargetMissing, events.InconsistentEventTypeNotEqual:
		var src T
		err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&src).Error
		switch {
		case err == nil:
			return f.target.WithContext(ctx).Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&src).Error
		case errors.Is(err, gorm.ErrRecordNotFound):
			return f.target.WithContext(ctx).Where("id = ?", src.ID()).Delete(new(T)).Error
		default:
			return err
		}
	case events.InconsistentEventBaseMissing:
		return f.target.WithContext(ctx).Where("id = ?", evt.Id).Delete(new(T)).Error
	default:
		return errors.New("未知数据不一致类型")
	}
}
