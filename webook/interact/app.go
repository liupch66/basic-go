package main

import (
	"github.com/liupch66/basic-go/webook/pkg/grpcx"
	"github.com/liupch66/basic-go/webook/pkg/saramax"
)

// 需要 main 函数控制启动、关闭的，在这都会有一个
type app struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
}
