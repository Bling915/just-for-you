package identityhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/identity"
	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type fakeLogin struct{ err error }

func (f fakeLogin) Execute(context.Context, application.LoginCommand) (application.Session, error) {
	return application.Session{}, f.err
}

type fakeRefresh struct{}

func (fakeRefresh) Execute(context.Context, application.RefreshSessionCommand) (application.Session, error) {
	return application.Session{}, domain.ErrInvalidRefreshToken
}

type fakeLogout struct{}

func (fakeLogout) Execute(context.Context, application.LogoutCommand) error { return nil }

type fakeGetProfile struct{}

func (fakeGetProfile) Execute(context.Context, application.GetProfileQuery) (domain.User, error) {
	return domain.NewUser(1, "open"), nil
}

type fakeUpdateProfile struct{}

func (fakeUpdateProfile) Execute(context.Context, application.UpdateProfileCommand) (domain.User, error) {
	return domain.User{}, domain.ErrInvalidNickname
}

func fakeUseCases(loginErr error) UseCases {
	return UseCases{
		Login: fakeLogin{err: loginErr}, RefreshSession: fakeRefresh{}, Logout: fakeLogout{},
		GetProfile: fakeGetProfile{}, UpdateProfile: fakeUpdateProfile{},
	}
}

func TestLoginKeepsStableErrorEnvelopeWithoutEchoingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(fakeUseCases(domain.ErrCodeRequired))
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
	NewHandler(fakeUseCases(nil)).RegisterProtected(group)
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", bytes.NewBufferString(`{"displayName":" ","avatarRef":""}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"INVALID_PROFILE"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
