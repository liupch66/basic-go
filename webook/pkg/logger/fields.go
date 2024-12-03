package logger

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Error(value error) Field {
	return Field{Key: "error", Value: value}
}
