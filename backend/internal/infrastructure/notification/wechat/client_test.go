package wechatnotification

import (
	"context"
	"encoding/json"
	application "github.com/nanfang/just-for-you/server/internal/application/companionship"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientMapsTemplatePayload(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "expires_in": 7200})
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["template_id"] != "tpl" || body["touser"] != "openid" {
			t.Fatalf("bad payload %#v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"errcode": 0})
	}))
	defer server.Close()
	c := NewClient(server.URL, "app", "secret", map[string]string{"moment": "tpl"})
	if e := c.Send(context.Background(), application.NotificationRecipient{OpenID: "openid"}, "moment", map[string]string{"thing1": "新动态"}); e != nil {
		t.Fatal(e)
	}
}
func TestClientNoopsWithoutTemplate(t *testing.T) {
	c := NewClient("http://invalid", "app", "secret", nil)
	if e := c.Send(context.Background(), application.NotificationRecipient{OpenID: "x"}, "moment", nil); e != nil {
		t.Fatal(e)
	}
}
