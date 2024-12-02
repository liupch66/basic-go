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
	// redis.Nil，它是一个特殊的错误值，表示 Redis 返回的是 空值，通常出现在对不存在的键执行 GET、HGET 等操作时，而不是 EXISTS。
	// 对于 Exists 命令，Result() 方法会返回两个值：
	// 1. 整数值：如果键存在，返回 1。如果键不存在，返回 0。
	// 2. 错误信息：如果命令执行成功，err 为 nil。如果发生错误（例如网络问题、Redis 服务不可用等），err 会包含错误信息。
	exists, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		// 下面这个不需要
		// if errors.Is(err, redis.Nil) {
		// 	return nil
		// }
		return err
	}
	if exists == 1 {
		return errors.New("用户已经退出登录")
	}
	return nil
}

func (h *RedisJwtHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segments := strings.Split(tokenHeader, " ")
	if len(segments) != 2 {
		return ""
	}
	return segments[1]
}
