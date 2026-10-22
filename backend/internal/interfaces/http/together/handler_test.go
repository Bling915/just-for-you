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

type handlerStorage struct{}

func (handlerStorage) SignUpload(context.Context, string, string, time.Time) (string, error) {
	return "https://upload", nil
}
func (handlerStorage) VerifyUploaded(context.Context, string, string, int64) error { return nil }
func (handlerStorage) SignRead(_ context.Context, key string, _ time.Time) (string, error) {
	return "https://read/" + key, nil
}

type handlerRepo struct{ item domain.Anniversary }

func (*handlerRepo) CreateCoverGrant(context.Context, app.CoverUploadGrant) error { return nil }
func (*handlerRepo) GetCoverGrant(context.Context, string) (app.CoverUploadGrant, error) {
	return app.CoverUploadGrant{}, domain.ErrInvalidCover
}
func (*handlerRepo) UpdateAnniversaryWithCoverGrant(_ context.Context, v domain.Anniversary, _ string, _ time.Time) (domain.Anniversary, error) {
	return v, nil
}

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
	NewHandler(app.NewServiceWithStorage(&handlerRepo{item: item}, handlerRelations{app.ActiveRelationship{ID: 9, UserAID: 2, UserBID: 3}}, handlerClock{now}, handlerStorage{}, func() string { return "key" })).RegisterProtected(g)
	return r
}

func TestPatchAnniversaryReturnsCompleteDTO(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "cover", "", "none", now, now)
	body := `{"title":"第一次旅行","date":"2026-11-01T00:00:00Z","countMode":"elapsed","repeatsYearly":false,"pinned":true,"coverRef":"cover","message":"出发"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/anniversaries/7", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	anniversaryRouter(item, now).ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"countMode":"elapsed"`) || !strings.Contains(res.Body.String(), `"pinned":true`) || !strings.Contains(res.Body.String(), `"message":"出发"`) || !strings.Contains(res.Body.String(), `"coverUrl":"https://read/cover"`) {
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

func TestDailyQuestionDTOHidesPartnerAnswerUntilUnlocked(t *testing.T) {
	now := time.Now()
	locked := domain.RestoreDailyQuestion(3, 9, now, "问题", "fallback", "v1", []domain.DailyAnswer{{UserID: 1, Text: "我的", SubmittedAt: now}, {UserID: 2, Text: "TA的", SubmittedAt: now}}, nil, now, now)
	dto := dailyQuestionDTO(locked, 1)
	if dto["partnerAnswer"] != nil {
		t.Fatalf("partner answer leaked: %+v", dto)
	}
	unlockedAt := now
	unlocked := domain.RestoreDailyQuestion(3, 9, now, "问题", "fallback", "v1", []domain.DailyAnswer{{UserID: 1, Text: "我的", SubmittedAt: now}, {UserID: 2, Text: "TA的", SubmittedAt: now}}, &unlockedAt, now, now)
	dto = dailyQuestionDTO(unlocked, 1)
	if dto["partnerAnswer"] == nil {
		t.Fatalf("partner answer missing: %+v", dto)
	}
}
