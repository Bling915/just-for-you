package togetherhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type handlerRelations struct{ rel app.ActiveRelationship }

func (h handlerRelations) GetActiveRelationship(context.Context, int64) (app.ActiveRelationship, error) {
	return h.rel, nil
}

type handlerClock struct{ now time.Time }

func (h handlerClock) Now() time.Time { return h.now }

type handlerRepo struct{ item domain.Anniversary }

func (h *handlerRepo) ListAnniversaries(context.Context, int64) ([]domain.Anniversary, error) {
	return []domain.Anniversary{h.item}, nil
}
func (h *handlerRepo) SaveAnniversary(_ context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	h.item = v
	return v, nil
}
func (h *handlerRepo) FindAnniversary(_ context.Context, coupleID, id int64) (domain.Anniversary, error) {
	if h.item.CoupleID() != coupleID || h.item.ID() != id {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	return h.item, nil
}
func (h *handlerRepo) UpdateAnniversary(_ context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	h.item = v
	return v, nil
}
func (h *handlerRepo) DeleteAnniversary(context.Context, int64, int64, int64) error { return nil }
func (*handlerRepo) ListTodos(context.Context, int64) ([]domain.Todo, error)        { return nil, nil }
func (*handlerRepo) SaveTodo(_ context.Context, v domain.Todo) (domain.Todo, error) { return v, nil }
func (*handlerRepo) FindTodo(context.Context, int64, int64) (domain.Todo, error) {
	return domain.Todo{}, nil
}
func (*handlerRepo) UpdateTodo(_ context.Context, v domain.Todo) (domain.Todo, error) { return v, nil }
func (*handlerRepo) DeleteTodo(context.Context, int64, int64, int64) error            { return nil }
func (*handlerRepo) ListWishes(context.Context, int64) ([]domain.Wish, error)         { return nil, nil }
func (*handlerRepo) SaveWish(_ context.Context, v domain.Wish) (domain.Wish, error)   { return v, nil }
func (*handlerRepo) FindWish(context.Context, int64, int64) (domain.Wish, error) {
	return domain.Wish{}, nil
}
func (*handlerRepo) UpdateWish(_ context.Context, v domain.Wish) (domain.Wish, error) { return v, nil }
func (*handlerRepo) DeleteWish(context.Context, int64, int64, int64) error            { return nil }
func (*handlerRepo) MomentCoupleID(context.Context, int64) (int64, error)             { return 9, nil }

func anniversaryRouter(item domain.Anniversary, now time.Time) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/v1", func(c *gin.Context) { c.Set("userID", int64(3)) })
	NewHandler(app.NewService(&handlerRepo{item: item}, handlerRelations{app.ActiveRelationship{ID: 9, UserAID: 2, UserBID: 3}}, handlerClock{now})).RegisterProtected(g)
	return r
}

func TestPatchAnniversaryReturnsCompleteDTO(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "", "", "none", now, now)
	body := `{"title":"第一次旅行","date":"2026-11-01T00:00:00Z","countMode":"elapsed","repeatsYearly":false,"pinned":true,"coverRef":"cover","message":"出发"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/anniversaries/7", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	anniversaryRouter(item, now).ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"countMode":"elapsed"`) || !strings.Contains(res.Body.String(), `"pinned":true`) || !strings.Contains(res.Body.String(), `"message":"出发"`) {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestPatchAnniversaryRejectsInvalidModeAndMessage(t *testing.T) {
	now := time.Now()
	item := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "", "", "none", now, now)
	for _, body := range []string{`{"title":"相识","date":"2026-11-01T00:00:00Z","countMode":"future"}`, `{"title":"相识","date":"2026-11-01T00:00:00Z","countMode":"elapsed","message":"` + strings.Repeat("念", 121) + `"}`} {
		req := httptest.NewRequest(http.MethodPatch, "/v1/anniversaries/7", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		anniversaryRouter(item, now).ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
		}
	}
}
