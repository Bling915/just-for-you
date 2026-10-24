package wish

import (
	"errors"
	"testing"
	"time"
)

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
