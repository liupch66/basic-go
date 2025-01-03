package fixer

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/liupch66/basic-go/webook/pkg/migrator"
)

// 这是后面重写的一个，和 fixer.go 基本差不多

type OverrideFixer[T migrator.Entity] struct {
	// 因为本身其实这个不涉及什么领域对象，
	// 这里操作的不是 migrator 本身的领域对象
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

// NewOverrideFixer 源表 src 和目标表 dst 不会变，但是 base 和 target 会变，direction 也会变
func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	// 在这里需要查询一下数据库中究竟有哪些列
	var t T
	rows, err := base.Model(&t).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (o *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var src T
	// 找出数据
	err := o.base.WithContext(ctx).Where("id = ?", id).First(&src).Error
	switch {
	// 找到了数据
	case err == nil:
		return o.target.Clauses(&clause.OnConflict{
			// 我们需要 Entity 告诉我们，修复哪些数据
			DoUpdates: clause.AssignmentColumns(o.columns),
		}).Create(&src).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		// DELETE FROM T.Table WHERE id = $id;
		return o.target.WithContext(ctx).Delete(new(T), id).Error
	default:
		return err
	}
}
