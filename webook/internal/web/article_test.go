package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/service"
	svcmocks "github.com/liupch66/basic-go/webook/internal/service/mocks"
	"github.com/liupch66/basic-go/webook/internal/web/jwt"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	loggermocks "github.com/liupch66/basic-go/webook/pkg/logger/mock"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1)
		// 这里 reqBody 也可以使用 web.ArticleReq,然后传入序列化的 reqBody,但是测试不了 bind 失败
		reqBody      []byte
		expectedCode int
		expectedRes  Result
	}{
		{
			name: "新建文章并发表成功",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1) {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "新建标题-->发表",
					Content: "新建内容-->发表",
					Author:  domain.Author{Id: 4},
				}).Return(int64(1), nil)
				return svc, nil
			},
			reqBody:      []byte(`{"title": "新建标题-->发表", "content": "新建内容-->发表"}`),
			expectedCode: 200,
			// 这里还可以和 article 的集成测试一样使用泛型
			expectedRes: Result{Msg: "OK", Data: float64(1)},
		},
		{
			name: "bind 失败",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1) {
				l := loggermocks.NewMockLoggerV1(ctrl)
				l.EXPECT().Info("article_publish bind 失败", logger.Error(errors.New("unexpected EOF")))
				return nil, l
			},
			reqBody:      []byte(`{" }`),
			expectedCode: 400,
			expectedRes:  Result{},
		},
		{
			name: "新建文章并发表失败",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "新建标题-->发表",
					Content: "新建内容-->发表",
					Author:  domain.Author{Id: 4},
				}).Return(int64(1), errors.New("mock publish error"))
				l.EXPECT().Info("发表文章失败", logger.Error(errors.New("mock publish error")))
				return svc, l
			},
			reqBody:      []byte(`{"title": "新建标题-->发表", "content": "新建内容-->发表"}`),
			expectedCode: 200,
			expectedRes:  Result{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewReader(tc.reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user_claims", jwt.UserClaims{UserId: 4})
			})
			articleHdl := NewArticleHandler(tc.mock(ctrl))
			articleHdl.RegisterRoutes(server)
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.expectedCode, resp.Code)
			if resp.Code == 200 {
				var res Result
				// Go 的 encoding/json 包默认将 JSON 数字解析为 float64 类型，即使原始 JSON 数据是整数
				err = json.NewDecoder(resp.Body).Decode(&res)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRes, res)
			}
		})
	}
}
