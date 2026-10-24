package companionship

import (
	"errors"
	"testing"
	"time"
)

func TestNewNudgeDerivesTheOtherCoupleMember(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	nudge, err := NewNudge(9, 7, 7, 8, NudgeHug, now)
	if err != nil {
		t.Fatal(err)
	}
	if nudge.SenderID() != 7 || nudge.RecipientID() != 8 || nudge.Kind() != NudgeHug {
		t.Fatalf("nudge=%+v", nudge)
	}
}

func TestNewNudgeRejectsUnknownKindAndNonMember(t *testing.T) {
	now := time.Now()
	if _, err := NewNudge(9, 7, 7, 8, NudgeKind("wave"), now); !errors.Is(err, ErrInvalidNudge) {
		t.Fatalf("unknown kind err=%v", err)
	}
	if _, err := NewNudge(9, 10, 7, 8, NudgeMissYou, now); !errors.Is(err, ErrNotCoupleMember) {
		t.Fatalf("non-member err=%v", err)
	}
}
