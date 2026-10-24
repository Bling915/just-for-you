package ai

import (
	"context"
	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	"testing"
	"time"
)

func TestFallbackQuestionIsDeterministicAndSafe(t *testing.T) {
	g := NewFallbackQuestions()
	c := app.QuestionContext{CoupleID: 9, Date: time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)}
	a, _ := g.Generate(context.Background(), c)
	b, _ := g.Generate(context.Background(), c)
	if a != b || a.Source != "fallback" || a.Text == "" {
		t.Fatalf("a=%+v b=%+v", a, b)
	}
}
