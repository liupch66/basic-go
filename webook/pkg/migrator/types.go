package migrator

type Entity interface {
	ID() int64
	CompareTo(dst Entity) bool
	// TableName() string
	// 这里要保证消息的有序性，并发时候用这个去修复是不行的，只能回去查
	// Columns() []string
}
