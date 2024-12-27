package ginx

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/liupch66/basic-go/webook/pkg/logger"
)

// L 使用包变量
var L logger.LoggerV1 = logger.NewNopLogger()

var vector *prometheus.CounterVec

func InitCounter(opts prometheus.CounterOpts) {
	// 这里 []string 也可以考虑设置 method, pattern, http status
	vector = prometheus.NewCounterVec(opts, []string{"code"})
	prometheus.MustRegister(vector)
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			L.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()), logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapReq[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		// 下半段的业务逻辑
		// 业务逻辑可能会操作 ctx, 比如读取 HTTP HEADER, 所以传入 ctx
		res, err := fn(ctx, req)
		if err != nil {
			L.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user_claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := fn(ctx, uc)
		if err != nil {
			L.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapReqAndClaims[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("user_claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := fn(ctx, req, uc)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			L.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
