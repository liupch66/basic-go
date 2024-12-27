package events

const (
	// InconsistentEventBaseMissing base 中没有数据
	InconsistentEventBaseMissing = "base_missing"

	// InconsistentEventTargetMissing target 中没有数据
	InconsistentEventTargetMissing = "target_missing"

	// InconsistentEventTypeNotEqual 目标表和源表的数据不相等
	InconsistentEventTypeNotEqual = "neq"
)

type InconsistentEvent struct {
	Id        int64
	Type      string
	Direction string
}
