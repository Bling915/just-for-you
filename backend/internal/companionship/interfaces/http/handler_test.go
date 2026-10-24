package companionshiphttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/companionship/application"
	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
)

type fakeUpdate struct{ err error }

func (f fakeUpdate) Execute(context.Context, application.UpdatePresenceCommand) (domain.Presence, error) {
	return domain.Presence{}, f.err
}

type fakeSend struct{ err error }

func (f fakeSend) Execute(context.Context, application.SendNudgeCommand) (domain.Nudge, error) {
	return domain.Nudge{}, f.err
}

type fakeSnapshot struct{}

func (fakeSnapshot) Execute(context.Context, application.GetSnapshotQuery) (application.Snapshot, error) {
	return application.Snapshot{}, nil
}

func TestPresenceMapsValidationErrors(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{{domain.ErrInvalidPresence, "INVALID_PRESENCE"}, {domain.ErrPresenceTextTooLong, "PRESENCE_TEXT_TOO_LONG"}, {domain.ErrInvalidPresenceExpiry, "INVALID_PRESENCE_EXPIRY"}}
	for _, item := range cases {
		t.Run(item.code, func(t *testing.T) {
			response := request(http.MethodPut, "/v1/presence", `{"preset":"custom"}`, Dependencies{UpdatePresence: fakeUpdate{item.err}, SendNudge: fakeSend{}, GetSnapshot: fakeSnapshot{}})
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"`+item.code+`"`) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestNudgeThrottleUsesStableErrorCode(t *testing.T) {
	response := request(http.MethodPost, "/v1/nudges", `{"kind":"hug"}`, Dependencies{UpdatePresence: fakeUpdate{}, SendNudge: fakeSend{domain.ErrNudgeThrottled}, GetSnapshot: fakeSnapshot{}})
	if response.Code != http.StatusTooManyRequests || !strings.Contains(response.Body.String(), `"code":"NUDGE_THROTTLED"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func request(method, path, body string, dependencies Dependencies) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(7)) })
	NewHandler(dependencies).RegisterProtected(group)
	response := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, req)
	return response
}

var _ = time.Time{}
