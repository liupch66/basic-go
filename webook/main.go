package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, world!")
	})
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}
