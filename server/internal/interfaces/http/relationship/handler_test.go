package relationshiphttp

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/relationship"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeCreate struct{}

func (fakeCreate) Execute(context.Context, application.CreateInvitationCommand) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}

type fakePreview struct{}

func (fakePreview) Execute(context.Context, application.PreviewInvitationQuery) (domain.Invitation, error) {
	return domain.Invitation{}, nil
}

type fakeBind struct{ err error }

func (f fakeBind) Execute(context.Context, application.BindRelationshipCommand) (domain.Relationship, error) {
	return domain.Relationship{}, f.err
}

type fakeGet struct{}

func (fakeGet) Execute(context.Context, application.GetRelationshipQuery) (domain.Relationship, error) {
	return domain.Relationship{}, domain.ErrNoActiveRelationship
}

type fakeEnd struct{}

func (fakeEnd) Execute(context.Context, application.EndRelationshipCommand) error { return nil }
func fakeUseCases(bindErr error) UseCases {
	return UseCases{fakeCreate{}, fakePreview{}, fakeBind{bindErr}, fakeGet{}, fakeEnd{}}
}
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
			NewHandler(fakeUseCases(item.err)).RegisterProtected(group)
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
	NewHandler(fakeUseCases(nil)).RegisterProtected(group)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/couples/me", nil))
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"NO_ACTIVE_COUPLE"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

var _ = errors.Is
