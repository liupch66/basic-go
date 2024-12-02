package logger

import (
	"go.uber.org/zap"
)

// ZapLogger 这里是用适配器模式
// 装饰器模式，一直都是同一个接口；而适配器模式，则必然是不同的接口
// 适配器模式：主要是为了将不兼容的接口连接起来，使其能够一起工作。它改变的是接口的结构。
// 装饰器模式：主要是为现有对象添加新功能，而不改变对象的结构。它改变的是对象的行为。
// 两者都涉及到包装现有对象，但适配器模式的目标是使不兼容的接口兼容，而装饰器模式的目标是增强对象的功能。
type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{l: l}
}

func (z *ZapLogger) toZapFields(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.l.Debug(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.l.Info(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.l.Warn(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.l.Error(msg, z.toZapFields(args)...)
}
