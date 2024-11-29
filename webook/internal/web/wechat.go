package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/service/oauth2/wechat"
)

var _ handler = (*OAuth2WechatHandler)(nil)

type OAuth2WechatHandler struct {
	wechatSvc wechat.Service
	userSvc   service.UserService
	jwtHandler
}

func NewOAuth2WechatHandler(wechatSvc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{wechatSvc: wechatSvc, userSvc: userSvc}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	{
		g.GET("/authurl", h.AuthURL)
		g.Any("/callback", h.Callback)
	}
	server.GET("/wechat/callback.do", h.YiHaoDian)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	url, err := h.wechatSvc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "构造微信扫码登录 URL 失败"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: url})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	info, err := h.wechatSvc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	u, err := h.userSvc.FindOrCreateByWechat(ctx, domain.WechatInfo{OpenId: info.OpenId, UnionId: info.UnionId})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	err = h.setJwtToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (h *OAuth2WechatHandler) YiHaoDian(ctx *gin.Context) {
	ctx.String(http.StatusOK, "欢迎来到一号店")
}
