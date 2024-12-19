package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"

	"github.com/liupch66/basic-go/webook/internal/service/sms"
)

type Service struct {
	svc sms.Service
	key []byte
}

func NewService(svc sms.Service, key []byte) sms.Service {
	return &Service{svc: svc, key: key}
}

type Claims struct {
	jwt.RegisteredClaims
	tplId string
}

// Send 使用 token 进行短信服务的权限控制,这里 tplToken 的语义发生了改变,之前就是 tplId, 这里是带 token 的
// 这个 tplToken 是线下申请的代表业务方的 token
func (s *Service) Send(ctx context.Context, tplToken string, params []string, numbers ...string) error {
	var c Claims
	token, err := jwt.ParseWithClaims(tplToken, &c, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("短信服务 token 不合法")
	}
	return s.svc.Send(ctx, tplToken, params, numbers...)
}
