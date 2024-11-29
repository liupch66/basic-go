package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId int64
	// 利用 UserAgent 增强登录安全性
	UserAgent string
}

func (h jwtHandler) setJwtToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}
