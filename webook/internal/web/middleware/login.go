package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (*LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	// 在 gin-session 中，存储非基础类型（如 time.Time 或自定义结构体）时，必须通过 gob.Register 注册。
	// 否则会报错：gob: type not registered for interface: time.Time
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		// 注册和登录不需要登录校验
		if ctx.Request.URL.Path == "/users/signup" || ctx.Request.URL.Path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 刷新 session
		now := time.Now()
		const timeKey = "update_time"
		updateTime := sess.Get(timeKey)
		updateTimeVal, ok := updateTime.(time.Time)
		// fmt.Printf("updateTime: %v, type: %T\n ", updateTime, updateTime)
		// 初次登录或超过 10s 刷新
		if updateTime == nil || (ok && now.Sub(updateTimeVal) > time.Second*10) {
			sess.Set(timeKey, now)
			sess.Options(sessions.Options{
				MaxAge: 60,
			})
			err := sess.Save()
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			fmt.Println("test")
		}
	}
}
