package article

import (
	"context"
	"math"
	"time"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

type ArticleReaderRepository interface {
	// Save 有就更新,没有就创建,即 upsert 的语义
	Save(ctx context.Context, art domain.Article) (int64, error)
}

// RetryableReaderRepositoryWithExponentialBackoff 尝试使用重试和指数退避的策略来执行 ReaderRepo 操作(装饰器模式)
type RetryableReaderRepositoryWithExponentialBackoff struct {
	repo     ArticleReaderRepository
	RetryMax int
	// 单位是秒
	MaxSleep float64
	l        logger.LoggerV1
}

func (r *RetryableReaderRepositoryWithExponentialBackoff) Save(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  int64
		err error
	)
	for i := 0; i < r.RetryMax; i++ {
		id, err = r.repo.Save(ctx, art)
		if err == nil {
			return id, nil
		}
		r.l.Error("保存到线上库失败", logger.Error(err), logger.Int("重试中,重试次数", i+1))
		sleepTime := math.Pow(2, float64(i))
		if sleepTime > r.MaxSleep {
			sleepTime = r.MaxSleep
		}
		time.Sleep(time.Second * time.Duration(sleepTime))
	}
	r.l.Error("保存到线上库重试全部失败", logger.Error(err))
	return 0, err
}
