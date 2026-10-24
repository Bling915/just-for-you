package todo

import (
	"testing"
	"time"
)

func TestTodoSharedCompletionAndCreatorDelete(t *testing.T) {
	now := time.Now()
	todo, _ := NewTodo(1, 2, "一起看电影", nil, now)
	done := todo.Complete(3, now)
	if done.Status() != "completed" || done.CompletedBy() == nil || *done.CompletedBy() != 3 {
		t.Fatal("partner should complete")
	}
	if done.CanDelete(3) || !done.CanDelete(2) {
		t.Fatal("only creator should delete")
	}
	if done.Reopen(now).Status() != "open" {
		t.Fatal("should reopen")
	}
}
