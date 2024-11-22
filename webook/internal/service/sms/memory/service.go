package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (svc *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	// 发送模板： {1}为您的登录验证码，请于{2}分钟内填写，如非本人操作，请忽略本短信。
	fmt.Println(params[0])
	return nil
}
