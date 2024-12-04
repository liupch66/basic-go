package service

import (
	"context"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
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
