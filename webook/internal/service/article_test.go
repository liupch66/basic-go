package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/repository/article"
	artRepomocks "github.com/liupch66/basic-go/webook/internal/repository/article/mocks"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	loggermocks "github.com/liupch66/basic-go/webook/pkg/logger/mock"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1)
		art  domain.Article

		expectedErr error
		expectedId  int64
	}{
		{
			name: "新建文章并发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				reader := artRepomocks.NewMockArticleReaderRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{Title: "新建标题", Content: "新建内容", Author: domain.Author{Id: 123}}).Return(int64(1), nil)
				// 注意制作库和线上库的 id 一致
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 1, Title: "新建标题", Content: "新建内容", Author: domain.Author{Id: 123}}).Return(int64(1), nil)
				return author, reader, nil
			},
			art:         domain.Article{Title: "新建标题", Content: "新建内容", Author: domain.Author{Id: 123}},
			expectedErr: nil,
			expectedId:  1,
		},
		{
			name: "修改已有文章-->保存到制作库-->发表(保存到线上库)成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				reader := artRepomocks.NewMockArticleReaderRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(2), nil)
				return author, reader, nil
			},
			art:         domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}},
			expectedErr: nil,
			expectedId:  2,
		},
		{
			name: "修改已有文章-->保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(errors.New("mock author repo error"))
				return author, nil, nil
			},
			art:         domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}},
			expectedErr: errors.New("mock author repo error"),
			expectedId:  0,
		},
		{
			name: "修改已有文章-->保存到制作库成功-->发表(保存到线上库)第一次失败-->重试一次成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				reader := artRepomocks.NewMockArticleReaderRepository(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 1))
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(2), nil)
				return author, reader, l
			},
			art:         domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}},
			expectedErr: nil,
			expectedId:  2,
		},
		{
			name: "修改已有文章-->保存到制作库成功-->发表(保存到线上库)第一次失败-->重试两次成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				reader := artRepomocks.NewMockArticleReaderRepository(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 1))
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 2))
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(2), nil)
				return author, reader, l
			},
			art:         domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}},
			expectedErr: nil,
			expectedId:  2,
		},
		{
			// 可以简化:不记录重试次数,直接使用 Times(3), 不管多少次就使用 AnyTimes
			name: "修改已有文章-->保存到制作库成功-->发表(保存到线上库)第一次失败-->重试三次(上限)全部失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository, logger.LoggerV1) {
				author := artRepomocks.NewMockArticleAuthorRepository(ctrl)
				reader := artRepomocks.NewMockArticleReaderRepository(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 1))
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 2))
				reader.EXPECT().Save(gomock.Any(), domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}}).Return(int64(0), errors.New("mock reader repo err"))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2), logger.Int("重试次数", 3))
				l.EXPECT().Error("保存到制作库成功-->保存到线上库全部重试失败", logger.Error(errors.New("mock reader repo err")),
					logger.Int64("article_id", 2))
				return author, reader, l
			},
			art:         domain.Article{Id: 2, Title: "修改标题", Content: "修改内容", Author: domain.Author{Id: 123}},
			expectedErr: errors.New("mock reader repo err"),
			expectedId:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := NewArticleServiceV1(tc.mock(ctrl))
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedId, id)
		})
	}
}
