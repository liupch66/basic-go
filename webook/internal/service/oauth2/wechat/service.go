package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"basic-go/webook/internal/domain"
	"basic-go/webook/pkg/logger"
)

const authURLPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redire"

// var redirectURL = url.PathEscape("https://meoying.com/oauth2/wechat/callback")
var redirectURL = `https%3A%2F%2Fpassport.yhd.com%2Fwechat%2Fcallback.do`

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewService(appId string, appSecret string, l logger.LoggerV1) Service {
	// 这里 client 不是严格的依赖注入
	return &service{appId: appId, appSecret: appSecret, client: http.DefaultClient, l: l}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	return fmt.Sprintf(authURLPattern, s.appId, redirectURL, state), nil
}

type Result struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`

	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	const targetPath = `https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`
	target := fmt.Sprintf(targetPath, s.appId, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	var res Result
	// 直接从 resp.Body（HTTP 响应体的流）中逐渐解码数据，而不需要将整个响应体先加载到内存中
	err = json.NewDecoder(resp.Body).Decode(&res)
	// 下面这个方法会读两遍：先读取整个响应体到内存中，然后进行解码 json 数据
	// body, err := io.ReadAll(resp.Body)
	// err = json.Unmarshal(body, &res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("微信返回了错误响应，错误码：%d，错误信息：%s", res.ErrCode, res.ErrMsg)
	}
	zap.L().Info("调用微信,拿到用户信息", zap.String("openId", res.OpenId), zap.String("unionId", res.UnionId))
	return domain.WechatInfo{OpenId: res.OpenId, UnionId: res.UnionId}, err
}
