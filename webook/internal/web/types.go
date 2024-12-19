package web

import (
	"github.com/gin-gonic/gin"

	"github.com/liupch66/basic-go/webook/pkg/ginx"
)

// Result 重构小技巧,用别名,不用去改其他代码
type Result = ginx.Result

type handler interface {
	RegisterRoutes(server *gin.Engine)
}
