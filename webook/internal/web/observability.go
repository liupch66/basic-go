package web

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ObservabilityHandler struct {
}

func NewObservabilityHandler() *ObservabilityHandler {
	return &ObservabilityHandler{}
}

func (o *ObservabilityHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/test")
	{
		g.GET("/metrics", func(ctx *gin.Context) {
			random := rand.Int31n(1000)
			time.Sleep(time.Duration(random) * time.Millisecond)
			ctx.String(http.StatusOK, "OK")
		})
	}
}
