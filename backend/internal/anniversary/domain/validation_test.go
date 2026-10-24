package anniversary

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
