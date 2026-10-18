package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthErrorDoesNotEchoSecrets(t *testing.T) {
	router := NewRouter(Dependencies{})
	secret := "sensitive-login-code"
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/wechat", bytes.NewBufferString(`{"code":"`+secret+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if strings.Contains(rec.Body.String(), secret) {
		t.Fatalf("response leaked code: %s", rec.Body.String())
	}
}
