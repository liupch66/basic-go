package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// AtKey access_token key
	AtKey = []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O")
	// RtKey refresh_token key
	RtKey = []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@1")
)

type RedisJwtHandler struct {
	cmd redis.Cmdable
}

func NewRedisJwtHandler(cmd redis.Cmdable) Handler {
	return &RedisJwtHandler{cmd: cmd}
}

func (h *RedisJwtHandler) SetJwtToken(ctx *gin.Context, userId int64, ssid string) error {
	uc := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) SetRefreshToken(ctx *gin.Context, userId int64, ssid string) error {
	rc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Ssid: ssid,
		Uid:  userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, rc)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	if err := h.SetJwtToken(ctx, uid, ssid); err != nil {
		return err
	}
	return h.SetRefreshToken(ctx, uid, ssid)
}

func (h *RedisJwtHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user_claims").(UserClaims)
	return h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", time.Hour*24*7).Err()
}

func (h *RedisJwtHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if logout > 0 {
		return errors.New("用户已经退出登录")
	}
	return err
}

func (h *RedisJwtHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segments := strings.Split(tokenHeader, " ")
	if len(segments) != 2 {
		return ""
	}
	return segments[1]
}
