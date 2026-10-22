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

func (fakeGetProfile) Execute(context.Context, application.GetProfileQuery) (application.Profile, error) {
	return application.Profile{User: domain.RestoreUser(1, "open", "小南", "users/1/avatars/a.jpg"), AvatarURL: "https://read/avatar"}, nil
}

type fakeUpdateProfile struct{}

func (fakeUpdateProfile) Execute(context.Context, application.UpdateProfileCommand) (application.Profile, error) {
	return application.Profile{}, domain.ErrInvalidNickname
}

type fakeAvatarUpload struct{ err error }

func (f fakeAvatarUpload) Execute(context.Context, application.RequestAvatarUploadCommand) (application.UploadGrant, error) {
	return application.UploadGrant{ObjectKey: "users/1/avatars/a.jpg", UploadURL: "https://upload/avatar", ContentType: "image/jpeg"}, f.err
}

func fakeUseCases(loginErr error) UseCases {
	return UseCases{
		Login: fakeLogin{err: loginErr}, RefreshSession: fakeRefresh{}, Logout: fakeLogout{},
		GetProfile: fakeGetProfile{}, UpdateProfile: fakeUpdateProfile{}, RequestAvatarUpload: fakeAvatarUpload{},
	}
}

func TestGetProfileReturnsStableRefAndSignedURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(1)) })
	NewHandler(fakeUseCases(nil)).RegisterProtected(group)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/users/me", nil))
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"avatarRef":"users/1/avatars/a.jpg"`) || !strings.Contains(body, `"avatarUrl":"https://read/avatar"`) {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
}

func TestAvatarUploadGrantValidatesRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(1)) })
	NewHandler(fakeUseCases(nil)).RegisterProtected(group)
	request := httptest.NewRequest(http.MethodPost, "/v1/users/me/avatar-upload", bytes.NewBufferString(`{"contentType":"image/jpeg","size":2048}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"uploadURL":"https://upload/avatar"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
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
