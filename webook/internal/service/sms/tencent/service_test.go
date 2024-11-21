package tencent

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func TestService_Send(t *testing.T) {
	// 这里我加入 wsl 的环境变量中结果查不到，可以编辑这个测试配置添加环境变量
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		t.Fatal()
	}
	number, ok := os.LookupEnv("NUMBER")
	if !ok {
		t.Fatal()
	}
	// region 查看：https://cloud.tencent.com/document/api/382/52071#.E5.9C.B0.E5.9F.9F.E5.88.97.E8.A1.A8
	client, err := sms.NewClient(&common.Credential{
		SecretId:  secretId,
		SecretKey: secretKey,
	}, "ap-nanjing", profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}
	s := NewService("1400859261", "webook公众号", client)

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string

		expectedErr error
	}{
		{
			name:        "发送验证码成功",
			tplId:       "1977183",
			params:      []string{"888", "5"},
			numbers:     []string{number},
			expectedErr: nil,
		},
		{
			name:        "发送验证码失败",
			tplId:       "1977183",
			params:      []string{"888", "5"},
			numbers:     []string{"10086"},
			expectedErr: nil,
		},
	}
	for _, tc := range testCases {
		err := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
		assert.Equal(t, tc.expectedErr, err)
	}
}
