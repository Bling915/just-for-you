package dailyquestionhttp

import (
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

func TestDailyQuestionDTOHidesPartnerAnswerUntilUnlocked(t *testing.T) {
	now := time.Now()
	locked := domain.RestoreDailyQuestion(
		3,
		9,
		now,
		"问题",
		"fallback",
		"v1",
		[]domain.DailyAnswer{
			{UserID: 1, Text: "我的", SubmittedAt: now},
			{UserID: 2, Text: "TA的", SubmittedAt: now},
		},
		nil,
		now,
		now,
	)
	dto := dailyQuestionDTO(locked, 1)
	if dto["partnerAnswer"] != nil {
		t.Fatalf("partner answer leaked: %+v", dto)
	}
	unlockedAt := now
	unlocked := domain.RestoreDailyQuestion(
		3,
		9,
		now,
		"问题",
		"fallback",
		"v1",
		[]domain.DailyAnswer{
			{UserID: 1, Text: "我的", SubmittedAt: now},
			{UserID: 2, Text: "TA的", SubmittedAt: now},
		},
		&unlockedAt,
		now,
		now,
	)
	dto = dailyQuestionDTO(unlocked, 1)
	if dto["partnerAnswer"] == nil {
		t.Fatalf("partner answer missing: %+v", dto)
	}
}
