package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	"github.com/liupch66/basic-go/webook/internal/domain"
	"github.com/liupch66/basic-go/webook/internal/service"
	"github.com/liupch66/basic-go/webook/internal/web/jwt"
	"github.com/liupch66/basic-go/webook/pkg/ginx"
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc interactv1.InteractServiceClient
	l        logger.LoggerV1
	biz      string
}

func NewArticleHandler(svc service.ArticleService, interSvc interactv1.InteractServiceClient, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{svc: svc, interSvc: interSvc, l: l, biz: "article"}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	{
		ag.POST("/edit", h.Edit)
		ag.POST("/publish", h.Publish)
		ag.POST("/withdraw", h.Withdraw)
		// 创作者的查询接口,查询列表页
		ag.POST("/list", ginx.WrapReqAndClaims[ListReq, jwt.UserClaims](h.List))
		ag.GET("/detail/:id", ginx.WrapClaims[jwt.UserClaims](h.Detail))
	}
	pg := server.Group("/pub")
	{
		pg.GET("/:id", ginx.WrapClaims[jwt.UserClaims](h.PubDetail))
		// 点赞和取消点赞都是这个
		pg.POST("/like", ginx.WrapReqAndClaims[LikeReq](h.Like))
		pg.POST("/collect", ginx.WrapReqAndClaims[CollectReq](h.Collect))
	}
}

// ReaderHandler 拆开比较合适，让 reader 服务于读者， 上面的 article 服务于作者，这里我们混在一起了
type ReaderHandler struct{}

func (h *ReaderHandler) RegisterRoutes(server *gin.Engine) {
	pg := server.Group("/pub")
	{
		pg.GET("/:id", h.PubDetail)
	}
}

func (h *ReaderHandler) PubDetail(ctx *gin.Context) {

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

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		h.l.Error("article_withdraw bind 失败", logger.Error(err))
		return
	}
	val, _ := ctx.Get("user_claims")
	uc, ok := val.(jwt.UserClaims)
	if !ok {
		h.l.Error("没有用户的 session 信息")
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	err := h.svc.Withdraw(ctx, req.toDomain(uc.UserId))
	if err != nil {
		h.l.Error("用户设置文章仅自己可见失败", logger.Error(err),
			logger.Int64("article_id", req.Id), logger.Int64("user_id", uc.UserId))
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "用户设置文章仅自己可见成功"})
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc jwt.UserClaims) (Result, error) {
	res, err := h.svc.List(ctx, uc.UserId, req.Offset, req.Limit)
	if err != nil {
		return Result{Code: 5, Msg: "系统错误"}, err
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Status:   src.Status.ToUnit8(),
				// 这个是创作者看自己的文章列表，不需要返回内容
				// Content: src.Content,
				// Author: src.Author
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc jwt.UserClaims) (Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Result{Code: 4, Msg: "参数错误"}, err
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		return Result{Code: 5, Msg: "系统错误"}, err
	}
	if art.Author.Id != uc.UserId {
		// 不需要告诉前端究竟发生了什么
		return Result{Code: 4, Msg: "输入有误"}, fmt.Errorf("非法访问文章,创作者 ID 不匹配: %d", uc.UserId)
	}
	return Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 查看文章详情就是需要 content，不需要 abstract
			// Abstract: art.Abstract(),
			Content: art.Content,
			Status:  art.Status.ToUnit8(),
			Ctime:   art.Ctime.Format(time.DateTime),
			Utime:   art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc jwt.UserClaims) (Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Result{Code: 4, Msg: "参数错误"}, fmt.Errorf("读者查询文章的 ID 错误，ID：%s， error：%w", idStr, err)
	}

	// 接下来查询文章详情和阅读点赞收藏详情可以并行，也就是开两个 goroutine，但是最终的 article_vo 要等这两个执行完
	// 所以可以用 WaitGroup，但是要各自处理错误，可以改进成 ErrGroup
	var (
		eg      errgroup.Group
		art     domain.Article
		getResp *interactv1.GetResponse
	)

	// goroutine 里面最好不要复用外面的 error，防止不清楚最后的 error 到底是哪个
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPublishedById(ctx, id, uc.UserId)
		return er
	})

	eg.Go(func() error {
		var er error
		getResp, er = h.interSvc.Get(ctx, &interactv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.UserId,
		})
		return er
	})

	err = eg.Wait()
	if err != nil {
		return Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("获取文章信息失败 %w", err)
	}

	// 增加阅读计数
	// 直接异步操作，在确定我们获取到了数据之后再来操作
	go func() {
		_, er := h.interSvc.IncrReadCnt(ctx, &interactv1.IncrReadCntRequest{
			Biz:   h.biz,
			BizId: id,
		})
		if er != nil {
			h.l.Error("增加阅读计数失败", logger.Error(er), logger.Int64("article_id", art.Id))
		}
	}()
	inter := getResp.Interact
	return Result{Data: ArticleVO{
		Id:    id,
		Title: art.Title,
		// Abstract: art.Abstract(),
		Content:    art.Content,
		Status:     art.Status.ToUnit8(),
		Author:     art.Author.Name,
		Liked:      inter.Liked,
		Collected:  inter.Collected,
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
		Ctime:      art.Ctime.Format(time.DateTime),
		Utime:      art.Utime.Format(time.DateTime),
	}}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc jwt.UserClaims) (Result, error) {
	var err error
	if req.Like {
		_, err = h.interSvc.Like(ctx, &interactv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	} else {
		_, err = h.interSvc.CancelLike(ctx, &interactv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	}
	if err != nil {
		h.l.Error("读者点赞或取消点赞失败", logger.Int64("article_id", req.Id),
			logger.Int64("user_id", uc.UserId), logger.Error(err))
		return Result{Code: 5, Msg: "系统错误"}, err
	}
	return Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc jwt.UserClaims) (Result, error) {
	_, err := h.interSvc.Collect(ctx, &interactv1.CollectRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Cid:   req.Cid,
		Uid:   uc.UserId,
	})
	if err != nil {
		h.l.Error("读者收藏失败", logger.Int64("article_id", req.Id),
			logger.Int64("user_id", uc.UserId), logger.Error(err))
		return Result{Code: 5, Msg: "系统错误"}, err
	}
	return Result{Msg: "OK"}, nil
}
