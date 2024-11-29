package domain

import (
	"time"
)

type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	// 考虑将来可能有 DingdingInfo，同名字段会冲突，所以这里没有组合
	WechatInfo WechatInfo
	Ctime      time.Time
}
