package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/integration/startup"
	"github.com/liupch66/basic-go/webook/internal/repository/dao/article"
	"github.com/liupch66/basic-go/webook/internal/web/jwt"
)

type ArticleGORMHandlerTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func TestArticleGORMHandler(t *testing.T) {
	suite.Run(t, &ArticleGORMHandlerTestSuite{})
}

func (a *ArticleGORMHandlerTestSuite) TestExample() {
	a.T().Log("测试套件")
}

// SetupTest: 在每个测试用例运行之前执行，适用于每个测试独立初始化
// SetupSuite: 在所有测试之前执行一次，适用于全局初始化
func (a *ArticleGORMHandlerTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(ctx *gin.Context) {
		ctx.Set("user_claims", jwt.UserClaims{UserId: 123})
		ctx.Next()
	})
	a.db = startup.InitTestDB()
	// articleHdl := startup.InitArticleHandler()
	articleHdl := startup.InitArticleHandler(article.NewGORMArticleDAO(a.db))
	articleHdl.RegisterRoutes(a.server)
}

// 大坑!!!坑爹啊!!!这个只会在 TestArticleHandler_Edit 最后运行一次,而不是每一个 tc 之后运行一次
// 这是在 ArticleGORMHandlerTestSuite 套件的每一个方法(例如这里的 TestExample 和 TestArticleHandler_Edit)之后运行一次,而不是每一个方法中的子测试之后
// func (a *ArticleGORMHandlerTestSuite) TearDownTest() {
// 	var count int64
// 	a.db.Model(&dao.Article{}).Count(&count)
// 	a.T().Log("当前数据库记录数:", count)
// 	a.T().Log("清理数据库")
// 	a.db.Exec("TRUNCATE TABLE articles")
// }

func (a *ArticleGORMHandlerTestSuite) TestArticleGORMHandler_Edit() {
	testCases := []struct {
		name         string
		before       func()
		after        func()
		art          Article
		expectedCode int
		expectedRes  Result[int64]
	}{
		{
			name:   "新建帖子-->保存成功(未发表)",
			before: func() {},
			after: func() {
				t := a.T()
				var art article.Article
				err := a.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				// 无法得知创建和更新准确时间
				now := time.Now().UnixMilli()
				assert.True(t, art.Ctime < now)
				assert.True(t, art.Ctime == art.Utime)
				// assert.True(t, req.Utime < now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "新建帖子",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUnit8(),
				}, art)
				t.Log("清理数据库")
				a.db.Exec("TRUNCATE TABLE articles")
			},
			art:          Article{Title: "新建帖子", Content: "新建内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Msg: "OK", Data: 1},
		},
		{
			name: "修改已有帖子(未发表)-->保存成功",
			before: func() {
				// 模拟已有帖子
				now := time.Now().UnixMilli()
				err := a.db.Create(article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUnit8(),
					Ctime:    now,
					Utime:    now,
				}).Error
				assert.NoError(a.T(), err)
			},
			after: func() {
				t := a.T()
				var art article.Article
				err := a.db.Where("id=?", 6).First(&art).Error
				assert.NoError(t, err)
				end := time.Now().UnixMilli()
				assert.True(t, art.Ctime < art.Utime)
				assert.True(t, art.Utime < end)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUnit8(),
				}, art)
				t.Log("清理数据库")
				a.db.Exec("TRUNCATE TABLE articles")
			},
			art:          Article{Id: 6, Title: "修改标题", Content: "修改内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Msg: "OK", Data: 6},
		},
		{
			name: "修改已有帖子(已发表)-->保存成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := a.db.Create(article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    now,
					Utime:    now,
				}).Error
				assert.NoError(a.T(), err)
			},
			after: func() {
				t := a.T()
				var art article.Article
				err := a.db.Where("id=?", 6).First(&art).Error
				assert.NoError(t, err)
				end := time.Now().UnixMilli()
				assert.True(t, art.Ctime < art.Utime)
				assert.True(t, art.Utime < end)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUnit8(),
				}, art)
				t.Log("清理数据库")
				a.db.Exec("TRUNCATE TABLE articles")
			},
			art:          Article{Id: 6, Title: "修改标题", Content: "修改内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Msg: "OK", Data: 6},
		},
		{
			name: "防止修改别人的帖子",
			before: func() {
				// 有一篇用户 234 的帖子,接下来用户 123 (user_claims 中的用户信息)想修改
				err := a.db.Create(article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 234,
					// 这里随便给个状态,验证没改变就行
					Status: domain.ArticleStatusPublished.ToUnit8(),
					Ctime:  8888,
					Utime:  8888,
				}).Error
				assert.NoError(a.T(), err)
			},
			after: func() {
				t := a.T()
				var art article.Article
				err := a.db.Where("id=?", 6).First(&art).Error
				assert.NoError(t, err)
				// 用户 123 肯定不能修改用户 234 的文章,所以 article 的任何信息都不会变,包括 ctime 和 utime
				assert.Equal(t, article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    8888,
					Utime:    8888,
				}, art)
				t.Log("清理数据库")
				a.db.Exec("TRUNCATE TABLE articles")
			},
			art:          Article{Id: 6, Title: "修改标题", Content: "修改内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Code: 5, Msg: "系统错误"},
		},
	}
	for _, tc := range testCases {
		a.Run(tc.name, func() {
			t := a.T()
			tc.before()
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			resp := httptest.NewRecorder()
			a.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.expectedCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedRes, res)
			tc.after()
		})
	}
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func (a *ArticleGORMHandlerTestSuite) TestArticleGORMHandler_Publish() {
	testCases := []struct {
		name string

		before func()
		after  func()

		art Article

		expectedCode int
		expectedRes  Result[int64]
	}{
		{
			name:   "新建文章并发表成功",
			before: func() {},
			after: func() {
				t := a.T()
				var art article.Article
				var pubArt article.PublishedArticle
				err := a.db.Where("id=? AND author_id=?", 1, 123).First(&art).Error
				assert.NoError(t, err)
				err = a.db.Where("id=? AND author_id=?", 1, 123).First(&pubArt).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime == art.Utime)
				assert.True(t, pubArt.Ctime == pubArt.Utime)
				assert.True(t, art.Ctime < pubArt.Ctime)
				assert.True(t, pubArt.Ctime < time.Now().UnixMilli())
				art.Ctime = 0
				art.Utime = 0
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
				}, art)
				assert.Equal(t, article.PublishedArticle{
					Id:       1,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
				}, pubArt)

				a.db.Exec("TRUNCATE TABLE articles")
				a.db.Exec("TRUNCATE TABLE published_articles")
			},
			art: Article{
				Title:   "新建标题",
				Content: "新建内容",
			},
			expectedCode: 200,
			expectedRes:  Result[int64]{Msg: "OK", Data: 1},
		},
		{
			// 制作库有,线上库没有
			name: "修改文章并第一次发表成功",
			before: func() {
				err := a.db.Create(&article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUnit8(),
					Ctime:    8888,
					Utime:    time.Now().UnixMilli(),
				}).Error
				assert.NoError(a.T(), err)
			},
			after: func() {
				t := a.T()
				var art article.Article
				var pubArt article.PublishedArticle
				err := a.db.Where("id=? AND author_id=?", 6, 123).First(&art).Error
				assert.NoError(t, err)
				err = a.db.Where("id=? AND author_id=?", 6, 123).First(&pubArt).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli()
				assert.True(t, art.Utime < now)
				assert.True(t, pubArt.Ctime == pubArt.Utime)
				assert.True(t, art.Utime < pubArt.Ctime)
				assert.True(t, pubArt.Ctime < now)
				art.Utime = 0
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, article.Article{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    8888,
				}, art)
				assert.Equal(t, article.PublishedArticle{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
				}, pubArt)

				a.db.Exec("TRUNCATE TABLE articles")
				a.db.Exec("TRUNCATE TABLE published_articles")
			},
			art: Article{
				Id:      6,
				Title:   "修改标题",
				Content: "修改内容",
			},
			expectedCode: 200,
			expectedRes:  Result[int64]{Msg: "OK", Data: 6},
		},
		{
			// 制作库有,线上库有
			name: "修改已发表文章并重新发表成功",
			before: func() {
				now := time.Now().UnixMilli()
				t := a.T()
				err := a.db.Create(&article.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    8888,
					Utime:    now,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&article.PublishedArticle{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    9999,
					Utime:    now,
				},
				).Error
				assert.NoError(t, err)
			},
			after: func() {
				t := a.T()
				var art article.Article
				var pubArt article.PublishedArticle
				err := a.db.Where("id=? AND author_id=?", 6, 123).First(&art).Error
				assert.NoError(t, err)
				err = a.db.Where("id=? AND author_id=?", 6, 123).First(&pubArt).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime < pubArt.Utime)
				assert.True(t, pubArt.Utime < time.Now().UnixMilli())
				art.Utime = 0
				pubArt.Utime = 0
				assert.Equal(t, article.Article{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    8888,
				}, art)
				assert.Equal(t, article.PublishedArticle{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUnit8(),
					Ctime:    9999,
				}, pubArt)

				a.db.Exec("TRUNCATE TABLE articles")
				a.db.Exec("TRUNCATE TABLE published_articles")
			},
			art: Article{
				Id:      6,
				Title:   "修改标题",
				Content: "修改内容",
			},
			expectedCode: 200,
			expectedRes:  Result[int64]{Msg: "OK", Data: 6},
		},
	}
	for _, tc := range testCases {
		a.Run(tc.name, func() {
			t := a.T()
			tc.before()
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			resp := httptest.NewRecorder()
			a.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.expectedCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedRes, res)
			tc.after()
		})
	}
}
