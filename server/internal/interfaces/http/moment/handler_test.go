package momenthttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeUpload struct{}

func (fakeUpload) Execute(context.Context, application.RequestPhotoUploadCommand) (application.UploadGrant, error) {
	return application.UploadGrant{}, domain.ErrInvalidUpload
}

type fakePublish struct{ err error }

func (f fakePublish) Execute(context.Context, application.PublishMomentCommand) (domain.Moment, error) {
	return domain.Moment{}, f.err
}

type fakeTimeline struct{}

func (fakeTimeline) Execute(context.Context, application.GetTimelineQuery) (application.TimelinePage, error) {
	return application.TimelinePage{}, nil
}

type fakeHome struct{}

func (fakeHome) Execute(context.Context, application.GetHomeQuery) (application.HomeResult, error) {
	return application.HomeResult{}, nil
}

type fakeDelete struct{ err error }

func (f fakeDelete) Execute(context.Context, application.DeleteMomentCommand) error { return f.err }
func fakeUseCases(publishErr, deleteErr error) UseCases {
	return UseCases{fakeUpload{}, fakePublish{publishErr}, fakeTimeline{}, fakeHome{}, fakeDelete{deleteErr}}
}
func routerFor(useCases UseCases) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(1)) })
	NewHandler(useCases).RegisterProtected(group)
	return router
}
func TestPublishMapsDomainErrors(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{{domain.ErrTooManyPhotos, "TOO_MANY_PHOTOS"}, {domain.ErrEmptyMoment, "EMPTY_MOMENT"}, {domain.ErrInvalidUpload, "INVALID_UPLOAD"}, {relationshipdomain.ErrNoActiveRelationship, "NO_ACTIVE_COUPLE"}}
	for _, item := range cases {
		t.Run(item.code, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/v1/moments", strings.NewReader(`{"text":"hello"}`))
			request.Header.Set("Content-Type", "application/json")
			routerFor(fakeUseCases(item.err, nil)).ServeHTTP(response, request)
			if !strings.Contains(response.Body.String(), `"code":"`+item.code+`"`) {
				t.Fatalf("body=%s", response.Body.String())
			}
		})
	}
}
func TestDeleteMapsNotAuthor(t *testing.T) {
	response := httptest.NewRecorder()
	routerFor(fakeUseCases(nil, domain.ErrNotAuthor)).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/v1/moments/9", nil))
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"NOT_AUTHOR"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
func TestPublishMapsInfrastructureFailureToServerError(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/moments", strings.NewReader(`{"text":"hello"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(fakeUseCases(errors.New("database unavailable"), nil)).ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
