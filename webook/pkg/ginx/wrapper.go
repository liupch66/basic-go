package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"basic-go/webook/pkg/logger"
)

// L 使用包变量
var L logger.LoggerV1

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func WrapBody[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
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

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
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

func WrapBodyAndToken[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
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
		if err != nil {
			L.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
