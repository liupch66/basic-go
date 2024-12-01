package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetJwtToken(ctx *gin.Context, userId int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, userId int64, ssid string) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractToken(ctx *gin.Context) string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Ssid   string
	UserId int64
	// 利用 UserAgent 增强登录安全性
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Ssid string
	Uid  int64
}
