package together

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"testing"
	"time"
)

type dailyRepoFake struct {
	value   domain.DailyQuestion
	created int
}

type dailyCoupleFake struct {
	relationship ActiveCouple
}

func (f dailyCoupleFake) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return f.relationship, nil
}

type dailyClockFake struct {
	now time.Time
}

func (f dailyClockFake) Now() time.Time {
	return f.now
}

func (f *dailyRepoFake) Find(context.Context, int64, time.Time) (domain.DailyQuestion, error) {
	if f.value.ID() == 0 {
		return domain.DailyQuestion{}, domain.ErrNotFound
	}
	return f.value, nil
}
func (f *dailyRepoFake) FindByID(context.Context, int64, int64) (domain.DailyQuestion, error) {
	return f.value, nil
}
func (f *dailyRepoFake) Create(_ context.Context, v domain.DailyQuestion) (domain.DailyQuestion, error) {
	f.created++
	f.value = domain.RestoreDailyQuestion(3, v.CoupleID(), v.Date(), v.Text(), v.Source(), v.ContextVersion(), nil, nil, v.CreatedAt(), v.UpdatedAt())
	return f.value, nil
}
func (f *dailyRepoFake) SubmitAnswer(context.Context, int64, int64, int64, string, time.Time) (domain.DailyQuestion, error) {
	return f.value, nil
}
func (f *dailyRepoFake) List(context.Context, int64, int) ([]domain.DailyQuestion, error) {
	return []domain.DailyQuestion{f.value}, nil
}

type generatorFake struct{ calls int }

func (g *generatorFake) Generate(context.Context, QuestionContext) (GeneratedQuestion, error) {
	g.calls++
	return GeneratedQuestion{Text: "今天最开心的小事？", Source: "fallback", ContextVersion: "v1"}, nil
}
func TestGetTodayUsesShanghaiDayAndCreatesOnlyOnce(t *testing.T) {
	now := time.Date(2026, 10, 21, 16, 30, 0, 0, time.UTC)
	repo := &dailyRepoFake{}
	generator := &generatorFake{}
	service := NewDailyQuestionService(
		dailyCoupleFake{relationship: ActiveCouple{ID: 9, UserAID: 1, UserBID: 2}},
		repo,
		generator,
		dailyClockFake{now: now},
	)
	first, err := service.GetToday(context.Background(), 1)
	if err != nil || first.Date().Day() != 22 || first.ContextVersion() != "v1" || generator.calls != 1 {
		t.Fatalf("first=%+v calls=%d err=%v", first, generator.calls, err)
	}
	_, err = service.GetToday(context.Background(), 1)
	if err != nil || generator.calls != 1 || repo.created != 1 {
		t.Fatalf("calls=%d created=%d err=%v", generator.calls, repo.created, err)
	}
}
