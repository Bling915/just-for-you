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
	"time"
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

type fakeUpdate struct {
	result domain.Moment
	err    error
}

func (f fakeUpdate) Execute(context.Context, application.UpdateMomentCommand) (domain.Moment, error) {
	return f.result, f.err
}
func fakeUseCases(publishErr, deleteErr error) UseCases {
	return UseCases{RequestPhotoUpload: fakeUpload{}, PublishMoment: fakePublish{publishErr}, UpdateMoment: fakeUpdate{}, GetTimeline: fakeTimeline{}, GetHome: fakeHome{}, DeleteMoment: fakeDelete{deleteErr}}
}
func TestUpdateMapsConflictAndReturnsOccurrenceVersion(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	conflict := fakeUseCases(nil, nil)
	conflict.UpdateMoment = fakeUpdate{err: domain.ErrMomentConflict}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/v1/moments/5", strings.NewReader(`{"text":"更新","occurredAt":"2026-10-20T11:00:00Z","version":"2026-10-20T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(conflict).ServeHTTP(response, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"MOMENT_CONFLICT"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}

	result := domain.RestoreMomentDetails(5, 1, 9, "更新", domain.StatusNone, "", nil, nil, nil, now.Add(-time.Hour), now.Add(-2*time.Hour), now, nil)
	success := fakeUseCases(nil, nil)
	success.UpdateMoment = fakeUpdate{result: result}
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/v1/moments/5", strings.NewReader(`{"text":"更新","occurredAt":"2026-10-20T11:00:00Z","version":"2026-10-20T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(success).ServeHTTP(response, request)
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"occurredAt":"2026-10-20T11:00:00Z"`) || !strings.Contains(body, `"updatedAt":"2026-10-20T12:00:00Z"`) {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
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
