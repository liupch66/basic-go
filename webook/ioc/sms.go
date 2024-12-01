package ioc

import (
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"basic-go/webook/internal/service/sms"
	"basic-go/webook/internal/service/sms/memory"
	"basic-go/webook/internal/service/sms/tencent"
)

func InitSmsService(cmd redis.Cmdable) sms.Service {
	// 装饰器模式,可以一直套
	// svc := ratelimit.NewService(memory.NewService(), limiter.NewRedisSlideWindowLimiter(cmd, 100, time.Second))
	// return retryable.NewService(svc, 3)
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
