package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"basic-go/webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(paths ...string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, paths...)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// if ctx.Request.URL.Path == "/users/signup" || ctx.Request.URL.Path == "/users/login" {
		// 	return
		// }
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.e30.HYOT9EjrWKoLsAeE4xv-nUprE4zHIu6C-X5a2cw0UD6UELNwbT-ViHMPemMz6uepSnxVeJcfRCKT_5iejyo82A
		segments := strings.Split(tokenHeader, " ")
		if len(segments) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var claims web.UserClaims
		token, err := jwt.ParseWithClaims(segments[1], &claims, func(*jwt.Token) (interface{}, error) {
			return []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"), nil
		})
		if err != nil || !token.Valid || claims.UserId == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// fmt.Println("current userAgent == last userAgent: ", claims.UserAgent == ctx.Request.UserAgent())
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题，要监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新 jwt
		if claims.ExpiresAt.Sub(time.Now()) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err := token.SignedString([]byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"))
			if err != nil {
				log.Println("jwt 续约失败：", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}

		// profile 接口可以复用 claims，不用再重新解析一遍
		ctx.Set("claims", claims)
	}
}
