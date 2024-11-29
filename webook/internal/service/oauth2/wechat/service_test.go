//go:build manual

package wechat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_service_manual_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	svc := NewService(appId, appSecret)
	// 从微信扫码那里拿一下
	res, err := svc.VerifyCode(context.Background(), "0238M6000sOXgT1imz2005q17c48M60S", "ApJxBSyDXKDJJFrhsEHmVr")
	require.NoError(t, err)
	t.Log(res)
}
