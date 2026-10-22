package ai

import (
	"context"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatibleValidatesOutputAndFallbacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("missing auth")
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"今天最想和对方分享哪件小事？"}}]}`))
	}))
	defer server.Close()
	client := NewOpenAICompatible(server.URL, "key", "model", server.Client())
	value, err := client.Generate(context.Background(), app.QuestionContext{})
	if err != nil || value.Source != "ai" {
		t.Fatalf("value=%+v err=%v", value, err)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"你是否患有抑郁症？"}}]}`))
	}))
	defer bad.Close()
	fallback := FallbackGenerator{Primary: NewOpenAICompatible(bad.URL, "key", "model", bad.Client()), Fallback: NewFallbackQuestions()}
	value, err = fallback.Generate(context.Background(), app.QuestionContext{})
	if err != nil || value.Source != "fallback" {
		t.Fatalf("value=%+v err=%v", value, err)
	}
}
