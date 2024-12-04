package service

import (
	"context"
	"time"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/article"
	"basic-go/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo article.ArticleRepository

	// V1
	authorRepo article.ArticleAuthorRepository
	readerRepo article.ArticleReaderRepository

	l logger.LoggerV1
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1) ArticleService {
	return &articleService{repo: repo, l: l}
}

func NewArticleServiceV1(authorRepo article.ArticleAuthorRepository, readerRepo article.ArticleReaderRepository,
	l logger.LoggerV1) ArticleService {
	return &articleService{authorRepo: authorRepo, readerRepo: readerRepo, l: l}
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := svc.repo.Update(ctx, art)
		return art.Id, err
	}
	return svc.repo.Create(ctx, art)
}

// 可以防止修改别人的帖子,但是性能较差,每次都要查询一次数据库
// 改进:在数据库更新帖子时的更新条件: where id = ? ----> where id = ? and author_id = ?
// func (svc *articleService) update(ctx context.Context, art domain.Article) error {
// 	res, err := svc.repo.FindById(ctx, art.Id)
// 	if err != nil {
// 		return err
// 	}
// 	if res.Author.Id != art.Author.Id {
// 		return errors.New("更改别人的帖子")
// 	}
// 	return svc.repo.Update(ctx, art)
// }

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	panic("implement me")
}

func (svc *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = svc.authorRepo.Update(ctx, art)
	} else {
		id, err = svc.authorRepo.Create(ctx, art)
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
		svc.l.Error("保存到制作库成功-->保存到线上库失败", logger.Error(err),
			logger.Int64("article_id", art.Id), logger.Int("重试次数", i+1))
		// 不一定要立马重试
		time.Sleep(time.Second * time.Duration(i))
	}
	svc.l.Error("保存到制作库成功-->保存到线上库全部重试失败", logger.Error(err), logger.Int64("article_id", art.Id))
	// 接入告警系统,人工处理
	// 走异步,保存到本地文件
	// 走 Canal
	// 打 MQ
	return 0, err
}
