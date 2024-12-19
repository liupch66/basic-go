package startup

import (
	"github.com/liupch66/basic-go/webook/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
