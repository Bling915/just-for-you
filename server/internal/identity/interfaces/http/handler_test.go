package identityhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/identity/application"
	"github.com/nanfang/just-for-you/server/internal/identity/domain"
)

type fakeUseCases struct{ loginErr error }

func (f fakeUseCases) LoginWithWeChat(context.Context, string) (application.Session, error) {
	return application.Session{}, f.loginErr
}
func (f fakeUseCases) RefreshSession(context.Context, string) (application.Session, error) {
	return application.Session{}, domain.ErrInvalidRefreshToken
}
func (f fakeUseCases) Logout(context.Context, string) error { return nil }
func (f fakeUseCases) GetMe(context.Context, int64) (domain.User, error) {
	return domain.NewUser(1, "open"), nil
}
func (f fakeUseCases) UpdateProfile(context.Context, int64, string, string) (domain.User, error) {
	return domain.User{}, domain.ErrInvalidNickname
}

func TestLoginKeepsStableErrorEnvelopeWithoutEchoingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(fakeUseCases{loginErr: domain.ErrCodeRequired})
	handler.RegisterPublic(router)
	secret := "sensitive-code"
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/wechat", bytes.NewBufferString(`{"code":"`+secret+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"LOGIN_FAILED"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), secret) {
		t.Fatalf("response leaked login code: %s", response.Body.String())
	}
}

func TestUpdateProfileMapsInvalidNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(1)) })
	NewHandler(fakeUseCases{}).RegisterProtected(group)
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", bytes.NewBufferString(`{"displayName":" ","avatarRef":""}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"INVALID_PROFILE"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
