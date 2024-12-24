package dao

import (
	"context"
	"errors"

	"go.uber.org/atomic"
)

// 双写的四个阶段
const (
	patternSrcOnly  = "SRC_ONLY"
	patternSrcFirst = "SRC_FIRST"
	patternDstFirst = "DST_FIRST"
	patternDstOnly  = "DST_ONLY"
)

var errUnknownPattern = errors.New("未知的双写 pattern")

type DoubleWriteDAO struct {
	src     InteractDAO
	dst     InteractDAO
	pattern *atomic.String
}

func NewDoubleWriteDAO(src InteractDAO, dst InteractDAO) *DoubleWriteDAO {
	// 默认就是 src 优先
	return &DoubleWriteDAO{
		src:     src,
		dst:     dst,
		pattern: atomic.NewString(patternSrcOnly),
	}
}

func (dao *DoubleWriteDAO) UpdatePattern(pattern string) {
	dao.pattern.Store(pattern)
}

// AST + 模版编程 = 代码生成

func (dao *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	switch dao.pattern.Load() {
	case patternSrcOnly:
		return dao.src.IncrReadCnt(ctx, biz, bizId)
	case patternSrcFirst:
		err := dao.src.IncrReadCnt(ctx, biz, bizId)
		// src 都出错了，不用考虑 dst
		if err != nil {
			return err
		}
		err = dao.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// dst 写失败无所谓，记日志，等校验与修复
		}
		return nil
	case patternDstFirst:
		err := dao.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = dao.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {

		}
		return nil
	case patternDstOnly:
		return dao.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return errUnknownPattern
	}
}

func (dao *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, biz string, bizId, cid, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interact, error) {
	switch dao.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return dao.src.Get(ctx, biz, bizId)
	case patternDstFirst, patternDstOnly:
		return dao.dst.Get(ctx, biz, bizId)
	default:
		return Interact{}, errUnknownPattern
	}
}

func (dao *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error {
	// TODO implement me
	panic("implement me")
}

func (dao *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interact, error) {
	// TODO implement me
	panic("implement me")
}
