package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	application "github.com/nanfang/just-for-you/server/internal/account/application"
)

type WeChatIdentityProvider struct {
	AppID, Secret string
	Client        *http.Client
}

func (w WeChatIdentityProvider) Exchange(ctx context.Context, code string) (application.ExternalIdentity, error) {
	query := url.Values{"appid": {w.AppID}, "secret": {w.Secret}, "js_code": {code}, "grant_type": {"authorization_code"}}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weixin.qq.com/sns/jscode2session?"+query.Encode(), nil)
	if err != nil {
		return application.ExternalIdentity{}, err
	}
	response, err := w.Client.Do(request)
	if err != nil {
		return application.ExternalIdentity{}, err
	}
	defer response.Body.Close()
	var payload struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrorCode  int    `json:"errcode"`
	}
	if err = json.NewDecoder(response.Body).Decode(&payload); err != nil || payload.ErrorCode != 0 || payload.OpenID == "" {
		return application.ExternalIdentity{}, fmt.Errorf("wechat exchange failed")
	}
	return application.ExternalIdentity{OpenID: payload.OpenID, SessionKey: payload.SessionKey}, nil
}
