package ginx

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
	Addr string
}

func NewServer(engine *gin.Engine, addr string) *Server {
	return &Server{Engine: engine, Addr: addr}
}

func (s *Server) Start() error {
	return s.Engine.Run(s.Addr)
}

// Result 可以通过在 Result 里面定义更加多的字段，来配合 Wrap 方法
// type Result struct {
// 	Code int    `json:"code"`
// 	Msg  string `json:"msg"`
// 	Data any    `json:"data"`
// }
//
// type UserClaims struct {
// 	Id        int64
// 	UserAgent string
// 	Ssid      string
// 	jwt.RegisteredClaims
// }
