package ratelimit

import (
	"context"
	"errors"
	"fmt"

	"github.com/liupch66/basic-go/webook/internal/service/sms"
	"github.com/liupch66/basic-go/webook/pkg/ratelimit"
)

const key = "sms_tencent"

var errLimited = errors.New("触发了限流")

// 装饰器模式的优点
// 增强功能：装饰器模式可以在不改变原有代码的情况下，增加新的功能（如限流、缓存、日志等）。
// 灵活组合：可以将多个装饰器组合在一起，形成更复杂的功能。每个装饰器只关注自己负责的功能。
// 开放/封闭原则：遵循“对扩展开放，对修改封闭”的设计原则，可以通过装饰器扩展服务的功能，而不需要修改已有代码。
// 侵入式修改会降低代码可读性,降低可测试性,强耦合,降低可拓展性

// Service Send 装饰器模式实现短信限流, svc 是被装饰者
// 可以有效阻止用户绕开装饰器
// 手动实现 sms.Service 接口所有方法
type Service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{svc: svc, limiter: limiter}
}

func (s *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return fmt.Errorf("短信服务判断限流出现错误: %w", err)
	}
	if limited {
		// 后面有需要的话,可以做成公开错误
		return errLimited
	}
	// 可以在这里加新代码,实现新特性
	// 没有侵入 svc.send 修改代码
	err = s.svc.Send(ctx, tplId, params, numbers...)
	// 也可以在这里加新代码,实现新特性
	return err
}

// ====================================================================================================

// ServiceV1 装饰器的另外一种实现方法:使用组合(composition)
// 自动实现 sms.Service 接口,可以只实现部分方法
// 用户可以绕开装饰器本身,直接访问 Service
type ServiceV1 struct {
	sms.Service
	limiter ratelimit.Limiter
}

func (s *ServiceV1) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return fmt.Errorf("短信服务判断限流出现错误: %w", err)
	}
	if limited {
		return errLimited
	}
	err = s.Service.Send(ctx, tplId, params, numbers...)
	// err = s.Send(ctx, tplId, params, numbers...)
	return err
}
