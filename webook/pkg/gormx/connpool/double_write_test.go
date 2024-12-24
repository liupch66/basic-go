package connpool

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestConnPool(t *testing.T) {
	src, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/webook"))
	require.NoError(t, err)
	err = src.AutoMigrate(&Interact{})
	require.NoError(t, err)

	dst, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/webook_interact"))
	require.NoError(t, err)
	err = dst.AutoMigrate(&Interact{})
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:    src.ConnPool,
			dst:    dst.ConnPool,
			patter: atomic.NewString(patternSrcFirst),
		},
	}))
	require.NoError(t, err)
	err = db.Create(&Interact{
		Biz:   "test",
		BizId: 1235,
	}).Error
	require.NoError(t, err)

	err = db.Transaction(func(tx *gorm.DB) error {
		return db.Create(&Interact{
			Biz:   "test_tx",
			BizId: 1235,
		}).Error
	})
	require.NoError(t, err)

	err = db.Model(&Interact{}).Where("id > ?", 0).Updates(map[string]any{
		"biz_id": 789,
	}).Error
	require.NoError(t, err)
}

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
