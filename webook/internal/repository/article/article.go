package article

import (
	"context"

	"gorm.io/gorm"

	"basic-go/webook/internal/domain"
	dao "basic-go/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步制作库和线上库数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO

	// V1 操作两个 DAO
	authorDao dao.AuthorDao
	readerDAO dao.ReaderDAO

	// V2 repository 层实现事务
	db *gorm.DB
}

func NewCachedArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{dao: dao}
}

func (repo *CachedArticleRepository) entityToDomain(ae dao.Article) domain.Article {
	return domain.Article{
		Id:      ae.Id,
		Title:   ae.Title,
		Content: ae.Content,
		Author:  domain.Author{Id: ae.AuthorId},
		Status:  domain.ArticleStatus(ae.Status),
	}
}

func (repo *CachedArticleRepository) domainToEntity(a domain.Article) (ae dao.Article) {
	return dao.Article{
		Id:       a.Id,
		Title:    a.Title,
		Content:  a.Content,
		AuthorId: a.Author.Id,
		Status:   a.Status.ToUnit8(),
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.domainToEntity(art))
}

func (repo *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.domainToEntity(art))
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Sync(ctx, repo.domainToEntity(art))
}

func (repo *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	a := repo.domainToEntity(art)
	// 先操作制作库
	if id == 0 {
		id, err = repo.authorDao.Insert(ctx, a)
	} else {
		err = repo.authorDao.Update(ctx, a)
	}
	if err != nil {
		return 0, err
	}
	// 再操作线上库,进行同步
	a.Id = id
	err = repo.readerDAO.Upsert(ctx, a)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止 panic 这里 defer 调用 rollback
	// 一旦你调用 Commit() 并且事务成功提交，后续的 Rollback() 将没有任何效果。
	defer tx.Rollback()
	authorDAO := dao.NewGORMAuthorDAO(tx)
	readerDAO := dao.NewGORMReaderDAO(tx)
	var (
		id  = art.Id
		err error
	)
	a := repo.domainToEntity(art)
	if id > 0 {
		err = authorDAO.Update(ctx, a)
	} else {
		id, err = authorDAO.Insert(ctx, a)
	}
	if err != nil {
		return 0, err
	}
	a.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(a))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, id, authorId, status.ToUnit8())
}
