package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"

	service2 "github.com/liupch66/basic-go/webook/interact/service"
	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository"
)

type RankService interface {
	TopN(ctx context.Context) error
	// TopN(ctx context.Context, n int64) error
	// TopN(ctx context.Context, n int64) ([]domain.Article, error)
}

type BatchRankService struct {
	artSvc    ArticleService
	interSvc  service2.InteractService
	repo      repository.RankRepository
	batchSize int
	n         int
	// 这里有个约束：不能返回负数
	scoreFunc func(utime time.Time, likeCnt int64) float64
}

func NewBatchRankService(artSvc ArticleService, interSvc service2.InteractService, repo repository.RankRepository) RankService {
	return &BatchRankService{
		artSvc:    artSvc,
		interSvc:  interSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(utime time.Time, likeCnt int64) float64 {
			// 这个 factor 也可以做成参数
			const factor = 1.5
			return float64(likeCnt-1) / math.Pow(time.Since(utime).Hours()+2, factor)
		},
	}
}

func (svc *BatchRankService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 这里存起来
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankService) topN(ctx context.Context) ([]domain.Article, error) {
	// 因为是分批次查询，要考虑耗时的影响，保证取的都是 now 之前的文章
	now := time.Now()
	offset := 0

	type Element struct {
		art   domain.Article
		score float64
	}
	que := queue.NewPriorityQueue[Element](svc.n, func(src Element, dst Element) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})

	for {
		// 先拿一批文章
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}

		// 这是刚好上一批取满而且刚好文章数据库取完，那还会进入这个批次。此时直接退出就行了
		if len(arts) == 0 {
			break
		}

		// 取文章 ids
		// ids := sliceMap[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
		// 	return src.Id
		// })
		ids := make([]int64, 0, len(arts))
		for _, art := range arts {
			ids = append(ids, art.Id)
		}

		// 根据文章 ids 再拿对应的点赞数据，这里拿到的结果是一个 map
		inters, err := svc.interSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}

		// 对取出来的一批排序
		for _, art := range arts {
			// inter, ok := inters[art.Id]
			// if !ok {
			// 	// 都没有点赞数据，肯定不是热度榜
			// 	continue
			// }
			ele := Element{art: art, score: svc.scoreFunc(art.Utime, inters[art.Id].LikeCnt)}
			err = que.Enqueue(ele)
			// topN 的 queue 已经满了
			if errors.Is(err, queue.ErrOutOfCapacity) {
				minEle, _ := que.Dequeue()
				if ele.score > minEle.score {
					_ = que.Enqueue(ele)
				} else {
					_ = que.Enqueue(minEle)
				}
			}
		}

		// 处理完这一批，要判断是否要进入下一批
		// 我这一批次都没取够，我肯定没有下一批了
		// 或者取到了七天前的数据，就不考虑放进热度榜了（取出来的 arts 排序是 utime DESC）
		ddl := now.Add(-7 * 24 * time.Hour)
		if len(arts) < svc.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}

		// 更新 offset 计数
		// offset += svc.batchSize
		offset += len(arts)
	}

	ql := que.Len()
	// 把热度榜 que 里面的 Article 取出来
	res := make([]domain.Article, ql)
	// 注意热度榜结果由高到低排列
	for i := ql - 1; i >= 0; i-- {
		ele, err := que.Dequeue()
		// 取完了，不够 ql
		if err != nil {
			break
		}
		res[i] = ele.art
	}
	return res, nil
}

func sliceMap[Src any, Dst any](src []Src, fn func(idx int, src Src) Dst) []Dst {
	dst := make([]Dst, len(src))
	for idx, s := range src {
		dst[idx] = fn(idx, s)
	}
	return dst
}
