package sms

import (
	"context"
)

type Service interface {
	// Send 适配不同服务商的抽象
	Send(ctx context.Context, tplId string, params []string, numbers ...string) error
}
