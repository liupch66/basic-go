package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/service/sms"
)

const codeTplId = "1977183"

var (
	ErrCodeSendTooMany   = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
	ErrCodeVerifyExpired = repository.ErrCodeVerifyExpired
)

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) generateCode() string {
	// 验证码不足 6 位时，用前导 0 填充为 6 位
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generateCode()
	// 塞进 redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送模板： {1}为您的登录验证码，请于{2}分钟内填写，如非本人操作，请忽略本短信。
	return svc.smsSvc.Send(ctx, codeTplId, []string{code, "10"}, phone)
}

func (svc *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		// 在接入了告警之后，这边要告警，防止有人搞你
		return false, nil
	}
	return ok, err
}
