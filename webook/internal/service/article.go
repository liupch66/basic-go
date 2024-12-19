package service

import (
	"context"
	"time"

	"github.com/liupch66/basic-go/webook/internal/domain"
	events "github.com/liupch66/basic-go/webook/internal/events/article"
	"github.com/liupch66/basic-go/webook/internal/repository/article"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

//go:generate mockgen -package=svcmocks -source=article.go -destination=mocks/article_mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
	// ListPub 因为是分批次查询，要考虑耗时的影响，保证取的都是 start 之前的文章
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo article.ArticleRepository

	// V1
	authorRepo          article.ArticleAuthorRepository
	readerRepo          article.ArticleReaderRepository
	retryableReaderRepo article.RetryableReaderRepositoryWithExponentialBackoff

	l        logger.LoggerV1
	producer events.Producer
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1, producer events.Producer) ArticleService {
	return &articleService{repo: repo, l: l, producer: producer}
}

func NewArticleServiceV1(authorRepo article.ArticleAuthorRepository, readerRepo article.ArticleReaderRepository,
	l logger.LoggerV1) ArticleService {
	return &articleService{authorRepo: authorRepo, readerRepo: readerRepo, l: l}
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// 保存到制作库,还未保存到线上库,也就未发表
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := svc.repo.Update(ctx, art)
		return art.Id, err
	}
	return svc.repo.Create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

func (svc *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = svc.authorRepo.Create(ctx, art)
	} else {
		err = svc.authorRepo.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 确保线上库和制作库的 id 一致
	art.Id = id
	// 防止部分失败,引入重试.也可以考虑装饰器模式:装饰 readerRepo
	for i := 0; i < 3; i++ {
		id, err = svc.readerRepo.Save(ctx, art)
		if err == nil {
			return id, nil
		}
		// 勘误:实际上有 err 记不到 id,懒得改了,要改还得一起改测试
		svc.l.Error("保存到制作库成功-->保存到线上库失败", logger.Error(err),
			logger.Int64("article_id", art.Id), logger.Int("重试次数", i+1))
		// 不一定要立马重试
		// 1.每次睡 1s (固定睡眠时间):简单、直观，但可能导致频繁请求和系统负载过重，适合错误恢复周期较短的场景。
		// time.Sleep(time.Second)
		// 2.每次睡递增的时间:避免请求过于频繁，适合需要恢复时间的场景，尤其是系统可能因为过载而失败的情况;
		// 但是,可能会导致过长的等待时间.比较复杂，可能需要调整增长的策略（如线性增长、指数增长等）
		time.Sleep(time.Second * time.Duration(i))
		// 3.指数退避(Exponential Backoff):通常会结合一个上限值，这意味着每次重试的时间不仅会增加，还会有一个最大限制。
		// 适合高失败率的场景，能够逐步减缓重试频率，并给系统更多恢复的时间;但是,如果指数退避过度，也可能导致请求的等待时间过长。
		// return svc.retryableReaderRepo.Save(ctx, art)
	}
	svc.l.Error("保存到制作库成功-->保存到线上库全部重试失败", logger.Error(err), logger.Int64("article_id", art.Id))
	// 接入告警系统,人工处理
	// 走异步,保存到本地文件
	// 走 Canal
	// 打 MQ
	return 0, err
}

func (svc *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	// 也可以设置 art.Status = domain.ArticleStatusPrivate,再接着往下传 art
	return svc.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (svc *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, uid, offset, limit)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := svc.repo.GetPublishedById(ctx, id)
	go func() {
		er := svc.producer.ProduceReadEvent(events.ReadEvent{Uid: uid, Aid: id})
		if er != nil {
			svc.l.Error("发送消息失败", logger.Int64("uid: ", uid), logger.Int64("aid: ", id), logger.Error(er))
		}
	}()
	return art, err
}

func (svc *articleService) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.ListPub(ctx, start, offset, limit)
}
