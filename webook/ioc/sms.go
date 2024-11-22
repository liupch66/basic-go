package ioc

import (
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"basic-go/webook/internal/service/sms"
	"basic-go/webook/internal/service/sms/memory"
	"basic-go/webook/internal/service/sms/tencent"
)

func InitSmsService() sms.Service {
	return memory.NewService()
}

func initSmsTencentService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_ID")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_KEY")
	}
	client, err := tencentSMS.NewClient(common.NewCredential(secretId, secretKey), "ap-nanjing", profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return tencent.NewService("1400859261", "webook公众号", client)
}
