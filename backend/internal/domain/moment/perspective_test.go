package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPerspectiveRequiresContentAndAllowsAtMostThreePhotos(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	value, err := NewPerspective(7, 2, "  我记得那天的风  ", []string{"a", "b", "c"}, now)
	if err != nil || value.Text() != "我记得那天的风" || len(value.Photos()) != 3 {
		t.Fatalf("value=%+v err=%v", value, err)
	}
	if _, err := NewPerspective(7, 2, "", []string{"a", "b", "c", "d"}, now); !errors.Is(err, ErrTooManyPerspectivePhotos) {
		t.Fatalf("err=%v", err)
	}
	if _, err := NewPerspective(7, 2, "", nil, now); !errors.Is(err, ErrEmptyPerspective) {
		t.Fatalf("err=%v", err)
	}
	if _, err := NewPerspective(7, 2, strings.Repeat("字", 501), nil, now); !errors.Is(err, ErrPerspectiveTextTooLong) {
		t.Fatalf("err=%v", err)
	}
}

func TestPerspectiveCanOnlyBeChangedByItsAuthor(t *testing.T) {
	value := RestorePerspective(3, 7, 2, "那天", nil, time.Now(), time.Now())
	if err := value.CanEdit(1); !errors.Is(err, ErrNotPerspectiveAuthor) {
		t.Fatalf("err=%v", err)
	}
	if err := value.CanEdit(2); err != nil {
		t.Fatalf("err=%v", err)
	}
}
