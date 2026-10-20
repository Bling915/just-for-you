package wechatnotification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	application "github.com/nanfang/just-for-you/server/internal/application/companionship"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	http                   *http.Client
	baseURL, appID, secret string
	templates              map[string]string
}

func NewClient(baseURL, appID, secret string, templates map[string]string) *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}, baseURL: baseURL, appID: appID, secret: secret, templates: templates}
}
func (c *Client) Send(ctx context.Context, to application.NotificationRecipient, scene string, data map[string]string) error {
	template := c.templates[scene]
	if template == "" || to.OpenID == "" {
		return nil
	}
	tokenURL := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", c.baseURL, url.QueryEscape(c.appID), url.QueryEscape(c.secret))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
	resp, e := c.http.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&token); e != nil {
		return e
	}
	if resp.StatusCode/100 != 2 || token.ErrCode != 0 || token.AccessToken == "" {
		return fmt.Errorf("wechat token: %d %s", token.ErrCode, token.ErrMsg)
	}
	values := map[string]map[string]string{}
	for k, v := range data {
		values[k] = map[string]string{"value": v}
	}
	body, _ := json.Marshal(map[string]any{"touser": to.OpenID, "template_id": template, "page": "pages/home/index", "data": values})
	sendURL := c.baseURL + "/cgi-bin/message/subscribe/send?access_token=" + url.QueryEscape(token.AccessToken)
	req, _ = http.NewRequestWithContext(ctx, http.MethodPost, sendURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, e = c.http.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return e
	}
	if resp.StatusCode/100 != 2 || result.ErrCode != 0 {
		return fmt.Errorf("wechat send: %d %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
