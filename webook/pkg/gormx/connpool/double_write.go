package connpool

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/atomic"
	"gorm.io/gorm"
)

// 双写的四个阶段
const (
	patternSrcOnly  = "SRC_ONLY"
	patternSrcFirst = "SRC_FIRST"
	patternDstFirst = "DST_FIRST"
	patternDstOnly  = "DST_ONLY"
)

var errUnknownPattern = errors.New("未知的双写 pattern")

type DoubleWritePool struct {
	src    gorm.ConnPool
	dst    gorm.ConnPool
	patter *atomic.String
}

func NewDoubleWritePool(src gorm.ConnPool, dst gorm.ConnPool) *DoubleWritePool {
	return &DoubleWritePool{
		src:    src,
		dst:    dst,
		patter: atomic.NewString(patternSrcOnly),
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) {
	d.patter.Store(pattern)
}

// 实现 GORM 的 ConnPool 接口

// PrepareContext 准备 SQL 语句，可以重复执行，并通过 *sql.Stmt（即预编译的 SQL 语句）提供 SQL 语句的执行接口。
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// sql.Stmt 是一个结构体，没有办法说返回一个代表双写的 Stmt
	// 禁用这个东西，因为我们没有办法创建出来 sql.Stmt 实例
	panic("implement me")
}

// ExecContext 执行没有返回结果的数据操作，如 INSERT、UPDATE、DELETE
func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.patter.Load() {
	case patternSrcOnly:
		// 虽然传入 args 编译器不报错，但是执行时 sql 会报错
		// sql: converting argument $1 type: unsupported type []interface {}, a slice of interface
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err := d.dst.ExecContext(ctx, query, args...)
			if err != nil {
				// 记日志，通知修复数据
			}
			return res, nil
		}
		return res, err
	case patternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err := d.src.ExecContext(ctx, query, args...)
			if err != nil {
				// 记日志，通知修复数据
			}
			return res, nil
		}
		return res, err
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryContext 执行查询并返回多行结果，返回 *sql.Rows，你可以通过它遍历所有查询结果。
func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.patter.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstFirst, patternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryRowContext 执行查询并返回单行结果，通常用于获取一条记录的数据。
func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.patter.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstFirst, patternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 因为返回值里面没有 error，只能 panic 掉
		panic(errUnknownPattern)
	}
}

// BeginTx 实现 GORM 事务接口（ConnPoolBeginner 或 TxBeginner），这里实现 ConnPoolBeginner
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.patter.Load()
	switch pattern {
	case patternSrcOnly:
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &DoubleWritePoolTx{
			src:    srcTx,
			patter: atomic.NewString(pattern),
		}, nil
	case patternSrcFirst:
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err == nil {
			dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
			if err != nil {
				// 记日志，通知修复数据
				// 也可以考虑回滚
			}
			return &DoubleWritePoolTx{
				src:    srcTx,
				dst:    dstTx,
				patter: atomic.NewString(pattern),
			}, nil
		}
		return nil, err
	case patternDstFirst:
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err == nil {
			srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
			if err != nil {
				// 记日志，通知修复数据
			}
			return &DoubleWritePoolTx{
				src:    srcTx,
				dst:    dstTx,
				patter: atomic.NewString(pattern),
			}, nil
		}
		return nil, err
	case patternDstOnly:
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &DoubleWritePoolTx{
			dst:    dstTx,
			patter: atomic.NewString(pattern),
		}, nil
	default:
		return nil, errUnknownPattern
	}
}

type DoubleWritePoolTx struct {
	src *sql.Tx
	dst *sql.Tx
	// 也可以用 string，因为事务不会并发
	patter *atomic.String
}

func (d *DoubleWritePoolTx) Commit() error {
	switch d.patter.Load() {
	case patternSrcOnly:
		return d.src.Commit()
	case patternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err := d.dst.Commit()
			if err != nil {
				// 记日志，通知修复数据
			}
		}
		return nil
	case patternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err := d.src.Commit()
			if err != nil {
				// 记日志，通知修复数据
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Commit()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.patter.Load() {
	case patternSrcOnly:
		return d.src.Rollback()
	case patternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err := d.dst.Rollback()
			if err != nil {
				// 记日志，通知修复数据
			}
		}
		return nil
	case patternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err := d.src.Rollback()
			if err != nil {
				// 记日志，通知修复数据
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Rollback()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// sql.Stmt 是一个结构体，没有办法说返回一个代表双写的 Stmt
	// 禁用这个东西，因为我们没有办法创建出来 sql.Stmt 实例
	panic("implement me")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.patter.Load() {
	case patternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			// 可能 dst 开事务失败了，不然会 panic
			if d.dst != nil {
				_, err := d.dst.ExecContext(ctx, query, args...)
				if err != nil {
					// 记日志，通知修复数据
				}
				return res, nil
			}
		}
		return res, err
	case patternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			if d.src != nil {
				_, err := d.src.ExecContext(ctx, query, args...)
				if err != nil {
					// 记日志，通知修复数据
				}
				return res, nil
			}
		}
		return res, err
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.patter.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstFirst, patternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.patter.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstFirst, patternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 因为返回值里面没有 error，只能 panic 掉
		panic(errUnknownPattern)
	}
}
