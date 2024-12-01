package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	ijwt "basic-go/webook/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{Handler: jwtHdl}
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
		tokenStr := l.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(*jwt.Token) (interface{}, error) {
			return []byte("C%B|]SiozBE,S)X>ru,3Uu0+rl1Lj.@O"), nil
		})
		if err != nil || !token.Valid || uc.UserId == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// fmt.Println("current userAgent == last userAgent: ", claims.UserAgent == ctx.Request.UserAgent())
		if uc.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题，要监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		if err = l.CheckSession(ctx, uc.Ssid); err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// profile 接口可以复用 claims，不用再重新解析一遍
		ctx.Set("user_claims", uc)
	}
}
