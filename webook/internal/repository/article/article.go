package article

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/cache"
	dao "basic-go/webook/internal/repository/dao/article"
	"basic-go/webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步制作库和线上库数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	userRepo repository.UserRepository
	dao      dao.ArticleDAO
	cache    cache.ArticleCache
	l        logger.LoggerV1

	// V1 操作两个 DAO
	authorDao dao.AuthorDao
	readerDAO dao.ReaderDAO

	// V2 repository 层实现事务
	db *gorm.DB
}

func NewCachedArticleRepository(dao dao.ArticleDAO, l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{dao: dao, l: l}
}

func (repo *CachedArticleRepository) entityToDomain(ae dao.Article) domain.Article {
	return domain.Article{
		Id:      ae.Id,
		Title:   ae.Title,
		Content: ae.Content,
		Author:  domain.Author{Id: ae.AuthorId},
		Status:  domain.ArticleStatus(ae.Status),
		Ctime:   time.UnixMilli(ae.Ctime),
		Utime:   time.UnixMilli(ae.Utime),
	}
}

func (repo *CachedArticleRepository) domainToEntity(a domain.Article) dao.Article {
	return dao.Article{
		Id:       a.Id,
		Title:    a.Title,
		Content:  a.Content,
		AuthorId: a.Author.Id,
		Status:   a.Status.ToUnit8(),
		Ctime:    a.Ctime.UnixMilli(),
		Utime:    a.Utime.UnixMilli(),
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		repo.cache.DeleteFirstPage(ctx, art.Author.Id)
	}()
	return repo.dao.Insert(ctx, repo.domainToEntity(art))
}

func (repo *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		repo.cache.DeleteFirstPage(ctx, art.Author.Id)
	}()
	return repo.dao.UpdateById(ctx, repo.domainToEntity(art))
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		repo.cache.DeleteFirstPage(ctx, art.Author.Id)
	}()
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

func (repo *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 1MB
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		err := repo.cache.Set(ctx, arts[0])
		if err != nil {
			repo.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func (repo *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 大部分情况下，没什么人会缓存分页的结果，因为如果数据的筛选条件、排序条件、分页的偏移量和数据量中的任何一个发生了变化，
	// 缓存就很难使用了.
	// 但是有一些缓存方案是可以考虑使用的。也就是只缓存第一页。在大多数的使用场景中，用户都是只看列表页的第一页，
	if offset == 0 && limit == 100 {
		data, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				// 查询完文章列表之后我猜很有可能访问第一篇文章，所以我选择预加载
				// 同时设置一个较短的过期时间，防止预测效果不好
				repo.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return repo.entityToDomain(src)
	})
	go func() {
		// 回写缓存
		// 不需要把整个 content 缓存下来,因为列表页也只展示了摘要
		err = repo.cache.SetFirstPage(ctx, uid, data)
		if err != nil {
			// 也不是什么大问题，可以输出 WARN
			repo.l.Warn("回写缓存失败", logger.Error(err))
		}
		repo.preCache(ctx, data)
	}()
	return data, nil
}

func (repo *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	ae, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.entityToDomain(ae), nil
}

func (repo *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	user, err := repo.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Ctime),
	}
	go func() {
		if err = repo.cache.Set(ctx, res); err != nil {
			repo.l.Warn("缓存已发表文章失败", logger.Error(err), logger.Int64("article_id", art.Id))
		}
	}()
	return res, nil
}
