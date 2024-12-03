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

func (a *ArticleTestSuite) TearDownTest() {
	a.db.Exec("TRUNCATE TABLE articles")
}

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
				assert.True(t, art.Utime < now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "新建帖子",
					Content:  "新建内容",
					AuthorId: 123,
				}, art)
			},
			art:          Article{Title: "新建帖子", Content: "新建内容"},
			expectedCode: http.StatusOK,
			expectedRes:  Result[int64]{Msg: "OK", Data: 1},
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
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
