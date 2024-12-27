package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/liupch66/basic-go/webook/pkg/migrator"
)

type Interact struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 联合索引的顺序：查询条件，区分度
	BizId      int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_id_type"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

func (i Interact) ID() int64 {
	return i.Id
}

func (i Interact) CompareTo(dst migrator.Entity) bool {
	dstVal, ok := dst.(Interact)
	return ok && i == dstVal
}

// UserLikeBiz 用户点赞表
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 三个构成唯一索引，前端展示点赞数的时候，查询条件：WHERE uid=? AND biz_id=? AND biz=?
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_id_type"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	// 依旧是只在 DB 层面生效的状态
	// 1- 有效，0-无效。软删除的用法
	Status uint8
	Ctime  int64
	Utime  int64
}

// Collection 收藏夹
type Collection struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(1024)"`
	Uid   int64
	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收藏夹 ID,作为关联关系中的外键，我们这里需要索引
	Cid int64 `gorm:"index"`
	// 查询条件：WHERE uid=? AND biz_id=? AND biz=?
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_id_type"`
	// 这算是一个冗余，因为正常来说，只需要在 Collection 中维持住 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:uid_biz_id_type"`
	Ctime int64
	Utime int64
}

//go:generate mockgen -package=mockdao -source=interact.go -destination=mocks/mock_interact.go InteractDAO
type InteractDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interact, error)
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error
	GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interact, error)
}

type GORMInteractDAO struct {
	db *gorm.DB
}

func NewGORMInteractDAO(db *gorm.DB) InteractDAO {
	return &GORMInteractDAO{db: db}
}

func (dao *GORMInteractDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	// check-do something 有并发问题
	// var inter Interact
	// err := dao.db.WithContext(ctx).Model(&Interact{}).Where("biz_id=? AND biz=?", bizId, biz).First(&inter).Error
	// if err != nil {
	// 	return err
	// }
	// err = dao.db.WithContext(ctx).Model(&Interact{}).Where("biz_id=? AND biz=?", bizId, biz).Updates(map[string]any{
	// 	"read_cnt": inter.ReadCnt+1,
	// 	"utime": now,
	// }).Error
	// return err

	// 可以开事务或者加锁，但是影响性能
	// 使用悲观锁锁定行 SELECT FOR UPDATE

	// 这里需要一个 upsert 的语义
	// 不需要开事务，利用 SQL 表达式就行，数据库会在执行更新时自动加锁该行记录并进行更新。
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		// MySQL 不写
		// Columns: []clause.Column{{Name: "biz_id"}, {Name: "biz"}},
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("read_cnt+1"),
			"utime":    now,
		}),
	}).Create(&Interact{
		BizId:   bizId,
		Biz:     biz,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
	// todo: 冲突时不是更新吗？为什么主键增加？WTF？
}

func (dao *GORMInteractDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新（点赞）表
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"status": 1,
				"utime":  now,
			}),
		}).Create(&UserLikeBiz{
			BizId:  bizId,
			Biz:    biz,
			Uid:    uid,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}
		// 更新（互动）表
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt+1"),
				"utime":    now,
			}),
		}).Create(&Interact{
			BizId:   bizId,
			Biz:     biz,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (dao *GORMInteractDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新（点赞）表
		err := tx.Model(&UserLikeBiz{}).Where("uid=? AND biz_id=? AND biz=?", uid, bizId, biz).
			Updates(map[string]any{
				"status": 0,
				"utime":  now,
			}).Error
		if err != nil {
			return err
		}
		// 更新（互动）表
		return tx.Model(&Interact{}).Where("biz_id=? AND biz=?", bizId, biz).
			Updates(map[string]any{
				"like_cnt": gorm.Expr("like_cnt-1"),
				"utime":    now,
			}).Error
	})
}

// InsertCollectionBiz 插入收藏记录，并更新计数
func (dao *GORMInteractDAO) InsertCollectionBiz(ctx context.Context, biz string, bizId, cid, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新收藏表
		err := tx.Create(&UserCollectionBiz{
			Cid:   cid,
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
			Ctime: now,
			Utime: now,
		}).Error
		if err != nil {
			return err
		}
		// 更新互动表
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("collect_cnt+1"),
				"utime":       now,
			}),
		}).Create(&Interact{
			BizId:      bizId,
			Biz:        biz,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

func (dao *GORMInteractDAO) Get(ctx context.Context, biz string, bizId int64) (Interact, error) {
	var res Interact
	err := dao.db.WithContext(ctx).Model(&Interact{}).Where("biz=? AND biz_id=?", biz, bizId).First(&res).Error
	return res, err
}

func (dao *GORMInteractDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=? AND status=?", uid, bizId, biz, 1).
		First(&res).Error
	return res, err
}

func (dao *GORMInteractDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", uid, bizId, biz).First(&res).Error
	return res, err
}

func (dao *GORMInteractDAO) BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractDAO(tx)
		for _, bizId := range bizIds {
			err := txDAO.IncrReadCnt(ctx, biz, bizId)
			if err != nil {
				// 也可以 return nil 容错记日志
				return err
			}
		}
		return nil
	})
}

func (dao *GORMInteractDAO) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interact, error) {
	var res []Interact
	err := dao.db.WithContext(ctx).Model(&Interact{}).Where("biz = ? AND biz_id IN ?", biz, bizIds).Find(&res).Error
	return res, err
}
