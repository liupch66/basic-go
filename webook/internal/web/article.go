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
		ag.POST("/publish", h.Publish)
	}
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uid},
	}
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		h.l.Error("article_bind失败", logger.Error(err))
		return
	}

	val := ctx.MustGet("user_claims")
	uc, ok := val.(jwt.UserClaims)
	if !ok {
		h.l.Error("没有用户的 session 信息")
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	id, err := h.svc.Save(ctx, req.toDomain(uc.UserId))
	// 不管什么 error,例如"修改别人的文章非法",都不会展示出来,统一"系统错误"
	if err != nil {
		h.l.Error("保存文章失败", logger.Error(err))
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK", Data: id})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		h.l.Info("article_publish bind 失败", logger.Error(err))
		return
	}
	// 没有这个值下一步也会断言出错,这里忽略
	val, _ := ctx.Get("user_claims")
	uc, ok := val.(jwt.UserClaims)
	if !ok {
		h.l.Info("没有用户的 session 信息")
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	// 有可能是新建然后直接发表,所以还是返回 id
	id, err := h.svc.Publish(ctx, req.toDomain(uc.UserId))
	if err != nil {
		h.l.Info("发表文章失败", logger.Error(err))
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK", Data: id})
}
