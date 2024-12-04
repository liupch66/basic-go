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

	"basic-go/webook/internal/integration/startup"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/web/jwt"
)

type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

func (a *ArticleTestSuite) TestExample() {
	a.T().Log("测试套件")
}

// SetupTest: 在每个测试用例运行之前执行，适用于每个测试独立初始化
// SetupSuite: 在所有测试之前执行一次，适用于全局初始化
func (a *ArticleTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(ctx *gin.Context) {
		ctx.Set("user_claims", jwt.UserClaims{UserId: 123})
	})
	articleHdl := startup.InitArticleHandler()
	articleHdl.RegisterRoutes(a.server)
	a.db = startup.InitTestDB()
}

// 大坑!!!坑爹啊!!!这个只会在 TestArticleHandler_Edit 最后运行一次,而不是每一个 tc 之后运行一次
// 这是在 ArticleTestSuite 套件的每一个方法(例如这里的 TestExample 和 TestArticleHandler_Edit)之后运行一次,而不是每一个方法中的子测试之后
// func (a *ArticleTestSuite) TearDownTest() {
// 	var count int64
// 	a.db.Model(&dao.Article{}).Count(&count)
// 	a.T().Log("当前数据库记录数:", count)
// 	a.T().Log("清理数据库")
// 	a.db.Exec("TRUNCATE TABLE articles")
// }

func (a *ArticleTestSuite) TestArticleHandler_Edit() {
	testCases := []struct {
		name         string
		before       func()
		after        func()
		art          Article
		expectedCode int
		expectedRes  Result[int64]
	}{
		{
			name:   "新建帖子-->保存成功",
			before: func() {},
			after: func() {
				t := a.T()
				var art dao.Article
				err := a.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				// 无法得知创建和更新准确时间
				now := time.Now().UnixMilli()
				assert.True(t, art.Ctime < now)
				assert.True(t, art.Ctime == art.Utime)
				// assert.True(t, art.Utime < now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "新建帖子",
					Content:  "新建内容",
					AuthorId: 123,
				}, art)
				t.Log("清理数据库")
				a.db.Exec("TRUNCATE TABLE articles")
			},
			art:          Article{Title: "新建帖子", Content: "新建内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Msg: "OK", Data: 1},
		},
		{
			name: "修改已有帖子-->保存成功",
			before: func() {
				// 模拟已有帖子
				a.db.Create(dao.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 123,
				})
			},
			after: func() {
				t := a.T()
				var art dao.Article
				err := a.db.Where("id=?", 6).First(&art).Error
				assert.NoError(t, err)
				end := time.Now().UnixMilli()
				assert.True(t, art.Ctime < art.Utime)
				assert.True(t, art.Utime < end)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       6,
					Title:    "修改标题",
					Content:  "修改内容",
					AuthorId: 123,
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
				a.db.Create(dao.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 234,
					Ctime:    8888,
					Utime:    8888,
				})
			},
			after: func() {
				t := a.T()
				var art dao.Article
				err := a.db.Where("id=?", 6).First(&art).Error
				assert.NoError(t, err)
				// 用户 123 肯定不能修改用户 234 的文章,所以 article 的任何信息都不会变,包括 ctime 和 utime
				assert.Equal(t, dao.Article{
					Id:       6,
					Title:    "新建标题",
					Content:  "新建内容",
					AuthorId: 234,
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