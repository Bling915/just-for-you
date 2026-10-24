package todo

import (
	"context"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type fakeTodoCouple struct{}

func (fakeTodoCouple) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return ActiveCouple{ID: 9}, nil
}

func TestRepeatedCompletionKeepsOriginalCompletedTime(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	created, err := domain.NewTodo(9, 7, "一起散步", nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryTodoStore{item: created}
	usecase := NewSetTodoCompleted(fakeTodoCouple{}, store, fixedTodoClock{now})

	first, err := usecase.Execute(context.Background(), SetTodoCompletedCommand{UserID: 7, ID: 1, Completed: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := usecase.Execute(context.Background(), SetTodoCompletedCommand{UserID: 8, ID: 1, Completed: true})
	if err != nil {
		t.Fatal(err)
	}
	if first.CompletedAt() == nil || !first.CompletedAt().Equal(*second.CompletedAt()) {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
}
