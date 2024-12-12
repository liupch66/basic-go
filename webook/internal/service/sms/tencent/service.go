package tencent

import (
	"context"
	"fmt"

	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
	// summaryVec *prometheus.SummaryVec // 接入记得初始化
}

func NewService(appId string, signName string, client *sms.Client) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	res := make([]*string, 0, len(src))
	for _, val := range src {
		res = append(res, &val)
	}
	return res
}

func (s *Service) Send(ctx context.Context, tplId string, params []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SetContext(ctx)
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tplId
	req.TemplateParamSet = s.toStringPtrSlice(params)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	resp, err := s.client.SendSms(req)
	zap.L().Debug("通过腾讯云发送短信", zap.Any("req", req), zap.Any("resp", resp), zap.Error(err))
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		// 这里想监控短信码只能侵入式监控了
		// code, _ := strconv.Atoi(*status.Code)
		// s.summaryVec.WithLabelValues().Observe(float64(code))
		//  空指针解引用会 panic
		if status == nil {
			return fmt.Errorf("短信发送失败")
		}
		if *status.Code != "Ok" {
			return fmt.Errorf("短信发送失败，错误码：%s，原因：%s", *status.Code, *status.Message)
		}
	}
	return nil
}
