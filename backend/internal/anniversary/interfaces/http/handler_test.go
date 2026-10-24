package anniversaryhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	app "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type handlerCouple struct{}

func (handlerCouple) GetActiveCouple(context.Context, int64) (app.ActiveCouple, error) {
	return app.ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}, nil
}

type handlerClock struct{}

func (handlerClock) Now() time.Time {
	return time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
}

type handlerStore struct {
	item domain.Anniversary
}

func (store *handlerStore) List(context.Context, int64) ([]domain.Anniversary, error) {
	return []domain.Anniversary{store.item}, nil
}

func (store *handlerStore) Save(
	_ context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	store.item = item
	return item, nil
}

func (store *handlerStore) Find(
	_ context.Context,
	coupleID, id int64,
) (domain.Anniversary, error) {
	if store.item.CoupleID() != coupleID || store.item.ID() != id {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	return store.item, nil
}

func (store *handlerStore) Update(
	_ context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	store.item = item
	return item, nil
}

func (*handlerStore) CreateCoverGrant(context.Context, app.CoverUploadGrant) error {
	return nil
}

func (*handlerStore) GetCoverGrant(context.Context, string) (app.CoverUploadGrant, error) {
	return app.CoverUploadGrant{}, domain.ErrInvalidCover
}

func (*handlerStore) UpdateWithCoverGrant(
	_ context.Context,
	item domain.Anniversary,
	_ string,
	_ time.Time,
) (domain.Anniversary, error) {
	return item, nil
}

func (*handlerStore) Delete(context.Context, int64, int64, int64) error {
	return nil
}

type handlerCoverURL struct{}

func (handlerCoverURL) Execute(_ context.Context, query app.CoverURLQuery) (string, error) {
	if query.CoverRef == "" {
		return "", nil
	}
	return "https://read/" + query.CoverRef, nil
}

func anniversaryRouter(item domain.Anniversary) *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := &handlerStore{item: item}
	relationships := handlerCouple{}
	handler := NewHandler(Dependencies{
		Update: app.NewUpdateAnniversary(
			relationships,
			store,
			nil,
			handlerClock{},
		),
		CoverURL: handlerCoverURL{},
	})
	router := gin.New()
	group := router.Group("/v1", func(c *gin.Context) {
		c.Set("userID", int64(3))
	})
	handler.RegisterProtected(group)
	return router
}

func TestPatchAnniversaryReturnsCompleteDTO(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(
		7,
		9,
		2,
		"相识",
		now,
		true,
		domain.CountModeCountdown,
		false,
		"cover",
		"",
		"none",
		now,
		now,
	)
	body := `{"title":"第一次旅行","date":"2026-11-01T00:00:00Z","countMode":"elapsed","repeatsYearly":false,"pinned":true,"coverRef":"cover","message":"出发"}`
	request := httptest.NewRequest(
		http.MethodPatch,
		"/v1/anniversaries/7",
		bytes.NewBufferString(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	anniversaryRouter(item).ServeHTTP(response, request)
	if response.Code != http.StatusOK ||
		!strings.Contains(response.Body.String(), `"countMode":"elapsed"`) ||
		!strings.Contains(response.Body.String(), `"pinned":true`) ||
		!strings.Contains(response.Body.String(), `"message":"出发"`) ||
		!strings.Contains(response.Body.String(), `"coverUrl":"https://read/cover"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPatchAnniversaryRejectsInvalidModeAndMessage(t *testing.T) {
	now := time.Now()
	item := domain.RestoreAnniversary(
		7,
		9,
		2,
		"相识",
		now,
		true,
		domain.CountModeCountdown,
		false,
		"",
		"",
		"none",
		now,
		now,
	)
	bodies := []string{
		`{"title":"相识","date":"2026-11-01T00:00:00Z","countMode":"future"}`,
		`{"title":"相识","date":"2026-11-01T00:00:00Z","countMode":"elapsed","message":"` +
			strings.Repeat("念", 121) + `"}`,
	}
	for _, body := range bodies {
		request := httptest.NewRequest(
			http.MethodPatch,
			"/v1/anniversaries/7",
			bytes.NewBufferString(body),
		)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		anniversaryRouter(item).ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	}
}
