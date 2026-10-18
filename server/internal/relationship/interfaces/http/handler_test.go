package relationshiphttp

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/relationship/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeUseCases struct{ bindErr error }

func (fakeUseCases) CreateInvitation(context.Context, int64) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}
func (fakeUseCases) PreviewInvitation(context.Context, int64, string) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}
func (f fakeUseCases) BindRelationship(context.Context, int64, string) (domain.Relationship, error) {
	return domain.Relationship{}, f.bindErr
}
func (fakeUseCases) GetRelationship(context.Context, int64) (domain.Relationship, error) {
	return domain.Relationship{}, domain.ErrNoActiveRelationship
}
func (fakeUseCases) EndRelationship(context.Context, int64) error { return nil }
func TestBindMapsStableErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{{domain.ErrInvitationExpired, "INVITE_EXPIRED"}, {domain.ErrInvitationUsed, "INVITE_USED"}, {domain.ErrCannotBindSelf, "CANNOT_BIND_SELF"}, {domain.ErrAlreadyBound, "ALREADY_BOUND"}}
	for _, item := range cases {
		t.Run(item.code, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(2)) })
			NewHandler(fakeUseCases{bindErr: item.err}).RegisterProtected(group)
			request := httptest.NewRequest(http.MethodPost, "/v1/couples/bind", bytes.NewBufferString(`{"code":"AB3K9Z"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if !strings.Contains(response.Body.String(), `"code":"`+item.code+`"`) {
				t.Fatalf("body=%s", response.Body.String())
			}
		})
	}
}
func TestGetMapsNoActiveRelationship(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(2)) })
	NewHandler(fakeUseCases{}).RegisterProtected(group)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/couples/me", nil))
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"NO_ACTIVE_COUPLE"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

var _ = errors.Is
