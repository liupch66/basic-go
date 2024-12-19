package ioc

import (
	"go.uber.org/zap"

	"github.com/liupch66/basic-go/webook/pkg/logger"
)

// func InitLogger() *zap.Logger {
// 	l, err := zap.NewDevelopment()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return l
// }

func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
