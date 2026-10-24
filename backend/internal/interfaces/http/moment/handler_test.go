package momenthttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
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

type fakePublish struct {
	err     error
	command application.PublishMomentCommand
}

func (f *fakePublish) Execute(_ context.Context, command application.PublishMomentCommand) (domain.Moment, error) {
	f.command = command
	return domain.Moment{}, f.err
}

type fakeTimeline struct{ query application.GetTimelineQuery }

func (f *fakeTimeline) Execute(_ context.Context, query application.GetTimelineQuery) (application.TimelinePage, error) {
	f.query = query
	return application.TimelinePage{}, nil
}

type fakeHome struct{}

func (fakeHome) Execute(context.Context, application.GetHomeQuery) (application.HomeResult, error) {
	return application.HomeResult{}, nil
}

type fakeDelete struct{ err error }

func (f fakeDelete) Execute(context.Context, application.DeleteMomentCommand) error { return f.err }

type fakeShare struct {
	command application.ShareToCoupleCommand
	result  domain.Moment
	err     error
}

func (f *fakeShare) Execute(_ context.Context, command application.ShareToCoupleCommand) (domain.Moment, error) {
	f.command = command
	return f.result, f.err
}

type fakeReact struct {
	command       application.ReactToMomentCommand
	deleteCommand application.DeleteReactionCommand
	result        domain.Reaction
	err           error
}
type fakePerspective struct {
	command       application.SavePerspectiveCommand
	deleteCommand application.DeletePerspectiveCommand
	result        domain.Perspective
	err           error
}
type fakeResurface struct {
	result application.ResurfacedMemoryResult
	err    error
}

func (f fakeResurface) Execute(context.Context, application.GetResurfacedMemoryQuery) (application.ResurfacedMemoryResult, error) {
	return f.result, f.err
}

func (f *fakePerspective) Execute(_ context.Context, command application.SavePerspectiveCommand) (domain.Perspective, error) {
	f.command = command
	return f.result, f.err
}
func (f *fakePerspective) Delete(_ context.Context, command application.DeletePerspectiveCommand) error {
	f.deleteCommand = command
	return f.err
}

func (f *fakeReact) Execute(_ context.Context, command application.ReactToMomentCommand) (domain.Reaction, error) {
	f.command = command
	return f.result, f.err
}

func (f *fakeReact) Delete(_ context.Context, command application.DeleteReactionCommand) error {
	f.deleteCommand = command
	return f.err
}

type fakeUpdate struct {
	result  domain.Moment
	err     error
	command application.UpdateMomentCommand
}

func (f *fakeUpdate) Execute(_ context.Context, command application.UpdateMomentCommand) (domain.Moment, error) {
	f.command = command
	return f.result, f.err
}
func fakeUseCases(publishErr, deleteErr error) UseCases {
	return UseCases{RequestPhotoUpload: fakeUpload{}, PublishMoment: &fakePublish{err: publishErr}, UpdateMoment: &fakeUpdate{}, GetTimeline: &fakeTimeline{}, GetHome: fakeHome{}, DeleteMoment: fakeDelete{deleteErr}, ShareToCouple: &fakeShare{}, ReactToMoment: &fakeReact{}, SavePerspective: &fakePerspective{}, GetResurfacedMemory: fakeResurface{}}
}
func TestResurfaceRouteReturnsEmptyMemory(t *testing.T) {
	response := httptest.NewRecorder()
	routerFor(fakeUseCases(nil, nil)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/memories/resurface", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"memory":null`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPerspectiveRoutesParseCommands(t *testing.T) {
	useCases := fakeUseCases(nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/v1/moments/7/perspective", strings.NewReader(`{"text":"我的视角","photoKeys":["p"]}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	command := useCases.SavePerspective.(*fakePerspective).command
	if response.Code != http.StatusOK || command.UserID != 1 || command.MomentID != 7 || command.Text != "我的视角" || len(command.PhotoKeys) != 1 {
		t.Fatalf("status=%d command=%+v body=%s", response.Code, command, response.Body.String())
	}
	response = httptest.NewRecorder()
	routerFor(useCases).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/v1/moments/7/perspective", nil))
	if response.Code != http.StatusNoContent || useCases.SavePerspective.(*fakePerspective).deleteCommand.MomentID != 7 {
		t.Fatalf("status=%d command=%+v", response.Code, useCases.SavePerspective.(*fakePerspective).deleteCommand)
	}
}

func TestShareAndReactionRoutesParseCommands(t *testing.T) {
	useCases := fakeUseCases(nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/moments/7/share-to-couple", strings.NewReader(`{"message":"给我们"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || useCases.ShareToCouple.(*fakeShare).command != (application.ShareToCoupleCommand{UserID: 1, MomentID: 7, Message: "给我们"}) {
		t.Fatalf("share status=%d command=%+v body=%s", response.Code, useCases.ShareToCouple.(*fakeShare).command, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/v1/moments/7/reactions", strings.NewReader(`{"type":"hug"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || useCases.ReactToMoment.(*fakeReact).command.Type != domain.ReactionHug {
		t.Fatalf("react status=%d command=%+v body=%s", response.Code, useCases.ReactToMoment.(*fakeReact).command, response.Body.String())
	}

	response = httptest.NewRecorder()
	routerFor(useCases).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/v1/moments/7/reactions/13", nil))
	if response.Code != http.StatusNoContent || useCases.ReactToMoment.(*fakeReact).deleteCommand != (application.DeleteReactionCommand{UserID: 1, MomentID: 7, ReactionID: 13}) {
		t.Fatalf("delete reaction status=%d command=%+v", response.Code, useCases.ReactToMoment.(*fakeReact).deleteCommand)
	}
}

func TestReactionValidationErrorsAreMapped(t *testing.T) {
	useCases := fakeUseCases(nil, nil)
	useCases.ReactToMoment = &fakeReact{err: domain.ErrReactionTextTooLong}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/moments/7/reactions", strings.NewReader(`{"type":"text","text":"too long"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"REACTION_TEXT_TOO_LONG"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
func TestUpdateMapsConflictAndReturnsOccurrenceVersion(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	conflict := fakeUseCases(nil, nil)
	conflict.UpdateMoment = &fakeUpdate{err: domain.ErrMomentConflict}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/v1/moments/5", strings.NewReader(`{"text":"更新","occurredAt":"2026-10-20T11:00:00Z","version":"2026-10-20T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(conflict).ServeHTTP(response, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"MOMENT_CONFLICT"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}

	result := domain.RestoreMomentDetails(domain.RestoreInput{ID: 5, AuthorID: 1, CoupleID: 9, Text: "更新", Status: domain.StatusNone, OccurredAt: now.Add(-time.Hour), CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now, SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, Kind: domain.KindStandard})
	success := fakeUseCases(nil, nil)
	success.UpdateMoment = &fakeUpdate{result: result}
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
	}{{domain.ErrTooManyPhotos, "TOO_MANY_PHOTOS"}, {domain.ErrEmptyMoment, "EMPTY_MOMENT"}, {domain.ErrInvalidUpload, "INVALID_UPLOAD"}, {coupledomain.ErrNoActiveCouple, "NO_ACTIVE_COUPLE"}}
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

func TestTimelineParsesSpaceAndOwner(t *testing.T) {
	useCases := fakeUseCases(nil, nil)
	timeline := useCases.GetTimeline.(*fakeTimeline)
	response := httptest.NewRecorder()
	routerFor(useCases).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/moments?space=personal&owner=partner&limit=12", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if timeline.query.SpaceType != domain.SpacePersonal || timeline.query.Owner != application.TimelineOwnerPartner || timeline.query.Limit != 12 {
		t.Fatalf("query=%+v", timeline.query)
	}
}

func TestTimelineRejectsInvalidSpaceQuery(t *testing.T) {
	response := httptest.NewRecorder()
	routerFor(fakeUseCases(nil, nil)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/moments?space=shared&owner=me", nil))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"INVALID_SPACE_QUERY"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPublishAndUpdateParseSpaceVisibility(t *testing.T) {
	useCases := fakeUseCases(nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/moments", strings.NewReader(`{"text":"我的小事","spaceType":"personal","visibility":"partner"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("publish status=%d body=%s", response.Code, response.Body.String())
	}
	publish := useCases.PublishMoment.(*fakePublish).command
	if publish.SpaceType != domain.SpacePersonal || publish.Visibility != domain.VisibilityPartner {
		t.Fatalf("publish=%+v", publish)
	}

	useCases = fakeUseCases(nil, nil)
	useCases.UpdateMoment = &fakeUpdate{result: domain.RestoreMomentDetails(domain.RestoreInput{ID: 5, AuthorID: 1, Text: "我的小事", SpaceType: domain.SpacePersonal, Visibility: domain.VisibilityPrivate, Kind: domain.KindStandard, OccurredAt: time.Date(2026, 10, 20, 11, 0, 0, 0, time.UTC), CreatedAt: time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)})}
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/v1/moments/5", strings.NewReader(`{"text":"我的小事","spaceType":"personal","visibility":"private","occurredAt":"2026-10-20T11:00:00Z","version":"2026-10-20T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	routerFor(useCases).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", response.Code, response.Body.String())
	}
	update := useCases.UpdateMoment.(*fakeUpdate).command
	if update.SpaceType != domain.SpacePersonal || update.Visibility != domain.VisibilityPrivate {
		t.Fatalf("update=%+v", update)
	}
}
