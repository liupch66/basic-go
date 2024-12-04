package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web/jwt"
	"basic-go/webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{svc: svc, l: l}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	{
		ag.POST("/edit", h.Edit)
	}
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		h.l.Error("article_bind失败", logger.Error(err))
		return
	}

	uc := ctx.MustGet("user_claims")
	val, ok := uc.(jwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		h.l.Error("没有用户的 session 信息")
		return
	}

	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: val.UserId},
	})
	// 不管什么 error,例如"修改别人的帖子非法",都不会展示出来,统一"系统错误"
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK", Data: id})
}
