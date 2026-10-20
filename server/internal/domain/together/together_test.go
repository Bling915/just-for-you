package together

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAnniversaryValidationAndLeapDay(t *testing.T) {
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if _, err := NewAnniversary(1, 2, strings.Repeat("纪", 61), now, now); !errors.Is(err, ErrInvalidTitle) {
		t.Fatalf("expected title error, got %v", err)
	}
	a, _ := NewAnniversary(1, 2, "相识", time.Date(2024, 2, 29, 0, 0, 0, 0, now.Location()), now)
	want := time.Date(2026, 2, 28, 0, 0, 0, 0, now.Location())
	if got := a.NextOccurrence(now); !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
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
func TestWishMomentMustBelongToCouple(t *testing.T) {
	now := time.Now()
	wish, _ := NewWish(7, 2, "去海边", "看日落", now)
	momentID := int64(9)
	if _, err := wish.Complete(3, &momentID, 8, now); !errors.Is(err, ErrInvalidMoment) {
		t.Fatalf("got %v", err)
	}
	done, err := wish.Complete(3, &momentID, 7, now)
	if err != nil || done.Status() != "completed" {
		t.Fatalf("completion failed: %v", err)
	}
}
