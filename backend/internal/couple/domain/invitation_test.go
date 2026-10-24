package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewInvitationRequiresSixUnambiguousCharacters(t *testing.T) {
	expires := time.Now().Add(time.Minute)
	for _, code := range []string{"ABC12", "ABC1234", "ABC10I", "abc123", "ABC-12"} {
		if _, err := NewInvitation(code, 1, expires); !errors.Is(err, ErrInvitationInvalid) {
			t.Fatalf("code=%q err=%v", code, err)
		}
	}
	if _, err := NewInvitation("AB3K9Z", 1, expires); err != nil {
		t.Fatal(err)
	}
}
func TestInvitationAcceptanceRules(t *testing.T) {
	now := time.Now()
	invitation, _ := NewInvitation("AB3K9Z", 1, now.Add(10*time.Minute))
	if err := invitation.CanBeAcceptedBy(1, now); !errors.Is(err, ErrCannotBindSelf) {
		t.Fatalf("self err=%v", err)
	}
	if err := invitation.CanBeAcceptedBy(2, now.Add(11*time.Minute)); !errors.Is(err, ErrInvitationExpired) {
		t.Fatalf("expiry err=%v", err)
	}
	invitation.MarkUsed(now, 2)
	if err := invitation.CanBeAcceptedBy(3, now); !errors.Is(err, ErrInvitationUsed) {
		t.Fatalf("used err=%v", err)
	}
}

func TestCouplePartnerOfRejectsNonMember(t *testing.T) {
	value := NewCouple(7, 1, 2, time.Now())
	if _, err := value.PartnerOf(3); !errors.Is(err, ErrNotCoupleMember) {
		t.Fatalf("PartnerOf() error = %v", err)
	}
}
