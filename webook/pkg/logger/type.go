package logger

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func Example() {
	var l Logger
	phone := "155XXXX5678"
	l.Info("手机用户未注册, phone: %s", phone)
}

type Field struct {
	Key   string
	Value any
}

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

func ExampleV1() {
	var l LoggerV1
	phone := "155XXXX5678"
	l.Info("手机用户未注册", Field{Key: "phone", Value: phone})
}

// LoggerV2 args 必须是偶数个,按照 key-value, key-value 组织下去
type LoggerV2 interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func ExampleV2() {
	var l LoggerV2
	phone := "155XXXX5678"
	l.Info("手机用户未注册", "phone", phone)
}
