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
	return svc.repo.Create(ctx, art)
}
