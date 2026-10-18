package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type HTTPWeChatExchanger struct {
	AppID, Secret string
	Client        *http.Client
}

func (w HTTPWeChatExchanger) Exchange(ctx context.Context, code string) (string, string, error) {
	u := "https://api.weixin.qq.com/sns/jscode2session?" + url.Values{"appid": {w.AppID}, "secret": {w.Secret}, "js_code": {code}, "grant_type": {"authorization_code"}}.Encode()
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	resp, e := w.Client.Do(req)
	if e != nil {
		return "", "", e
	}
	defer resp.Body.Close()
	var x struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&x); e != nil || x.ErrCode != 0 || x.OpenID == "" {
		return "", "", fmt.Errorf("wechat exchange failed")
	}
	return x.OpenID, x.SessionKey, nil
}
