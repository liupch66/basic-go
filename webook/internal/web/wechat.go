package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lithammer/shortuuid/v4"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/service/oauth2/wechat"
	ijwt "basic-go/webook/internal/web/jwt"
)

var _ handler = (*OAuth2WechatHandler)(nil)

type OAuth2WechatHandler struct {
	wechatSvc wechat.Service
	userSvc   service.UserService
	stateKey  []byte
	cfg       WechatHandlerConfig
	ijwt.Handler
}

// WechatHandlerConfig 生产环境设置为 true
type WechatHandlerConfig struct {
	Secure bool
}

func NewOAuth2WechatHandler(wechatSvc wechat.Service, userSvc service.UserService, cfg WechatHandlerConfig, jwtHdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		wechatSvc: wechatSvc,
		userSvc:   userSvc,
		stateKey:  []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"),
		cfg:       cfg,
		Handler:   jwtHdl,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	{
		g.GET("/authurl", h.AuthURL)
		g.Any("/callback", h.Callback)
	}
	server.GET("/wechat/callback.do", h.YiHaoDian)
}

type StateClaim struct {
	jwt.RegisteredClaims
	State string
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := shortuuid.New()
	url, err := h.wechatSvc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "构造微信扫码登录 URL 失败"})
		return
	}
	if err = h.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Data: url})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	sc := StateClaim{
		State:            state,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute))},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, sc)
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback", "", h.cfg.Secure, true)
	return nil
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "微信扫码登录失败"})
		return
	}

	info, err := h.wechatSvc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	u, err := h.userSvc.FindOrCreateByWechat(ctx, domain.WechatInfo{OpenId: info.OpenId, UnionId: info.UnionId})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if err = h.SetLoginToken(ctx, u.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	// 检验 state
	tokenStr, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("拿不到 state 的 cookie, %w", err)
	}
	var sc StateClaim
	token, err := jwt.ParseWithClaims(tokenStr, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("cookie 不是合法 jwt token, %w", err)
	}
	if sc.State != state {
		return errors.New("state 被篡改了")
	}
	return nil
}

func (h *OAuth2WechatHandler) YiHaoDian(ctx *gin.Context) {
	ctx.String(http.StatusOK, "欢迎来到一号店")
}
