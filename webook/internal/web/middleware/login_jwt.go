package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct{}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (*LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" || ctx.Request.URL.Path == "/users/login" {
			return
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
		token, err := jwt.Parse(segments[1], func(*jwt.Token) (interface{}, error) {
			return []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
