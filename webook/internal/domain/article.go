package domain

import (
	"time"
)

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	// 摘要我们取前几句。要考虑一个中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
}

type Author struct {
	Id   int64
	Name string
}

// ArticleStatus 状态码, 这是 uint8 的衍生类型,或叫自定义类型
// 类型别名: type byte = uint8,type rune = int32
type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题,有状态的状态码尽量不要用 0
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUnit8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) Valid() bool {
	// Go 允许自定义类型和基础类型之间进行比较
	return s > 0
}

func (s ArticleStatus) Published() bool {
	return s == ArticleStatusPublished
}

// String 每加一个 status 就要加一个 case, 可以考虑用 V1
func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	case ArticleStatusPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// ArticleStatusV1 如果有很多方法,状态里面需要一些额外字段,就用这个版本
type ArticleStatusV1 struct {
	Val  uint8
	Name string
}

var (
	ArticleStatusV1Unknown     = ArticleStatusV1{Val: 0, Name: "unknown"}
	ArticleStatusV1Unpublished = ArticleStatusV1{Val: 1, Name: "unpublished"}
	ArticleStatusV1Published   = ArticleStatusV1{Val: 2, Name: "published"}
	ArticleStatusV1Private     = ArticleStatusV1{Val: 3, Name: "private"}
)

func (s ArticleStatusV1) String() string {
	return s.Name
}

// ArticleStatusV2 出于可读性的考虑也可以这样定义
type ArticleStatusV2 string
