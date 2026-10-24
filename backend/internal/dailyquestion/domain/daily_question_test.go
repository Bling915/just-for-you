package dailyquestion

import (
	"errors"
	"testing"
	"time"
)

func TestDailyQuestionUnlocksOnlyAfterBothAnswersAndThenFreezes(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	q, err := NewDailyQuestion(9, now, "今天哪件小事让你开心？", "fallback", now)
	if err != nil {
		t.Fatal(err)
	}
	q, unlocked, err := q.Submit(1, "一杯热咖啡", 2, now)
	if err != nil || unlocked || q.Unlocked() {
		t.Fatalf("q=%+v unlocked=%v err=%v", q, unlocked, err)
	}
	q, unlocked, err = q.Submit(2, "一起散步", 1, now.Add(time.Minute))
	if err != nil || !unlocked || !q.Unlocked() {
		t.Fatalf("q=%+v unlocked=%v err=%v", q, unlocked, err)
	}
	if _, _, err = q.Submit(1, "修改", 2, now.Add(time.Hour)); !errors.Is(err, ErrDailyQuestionFrozen) {
		t.Fatalf("err=%v", err)
	}
}
func TestDailyAnswerValidatesText(t *testing.T) {
	q, _ := NewDailyQuestion(9, time.Now(), "问题", "fallback", time.Now())
	if _, _, err := q.Submit(1, "   ", 2, time.Now()); !errors.Is(err, ErrInvalidDailyAnswer) {
		t.Fatalf("err=%v", err)
	}
}
