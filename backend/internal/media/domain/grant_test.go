package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewGrantValidatesPurposeOwnerResourceAndImage(t *testing.T) {
	now := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	valid := []GrantInput{
		{ObjectKey: "avatars/a", OwnerID: 1, Purpose: PurposeAvatar, ContentType: "image/jpeg", Size: 1, ExpiresAt: now.Add(time.Minute)},
		{ObjectKey: "moments/a", OwnerID: 1, Purpose: PurposeMomentPhoto, ContentType: "image/png", Size: 10 << 20, ExpiresAt: now.Add(time.Minute)},
		{ObjectKey: "covers/a", OwnerID: 1, ResourceID: 9, Purpose: PurposeAnniversaryCover, ContentType: "image/webp", Size: 5 << 20, ExpiresAt: now.Add(time.Minute)},
	}
	for _, input := range valid {
		if _, err := NewGrant(input, now); err != nil {
			t.Fatalf("input=%+v err=%v", input, err)
		}
	}
	invalid := []GrantInput{
		{ObjectKey: "x", OwnerID: 0, Purpose: PurposeAvatar, ContentType: "image/jpeg", Size: 1, ExpiresAt: now.Add(time.Minute)},
		{ObjectKey: "x", OwnerID: 1, Purpose: PurposeAnniversaryCover, ContentType: "image/jpeg", Size: 1, ExpiresAt: now.Add(time.Minute)},
		{ObjectKey: "x", OwnerID: 1, Purpose: PurposeAvatar, ContentType: "text/plain", Size: 1, ExpiresAt: now.Add(time.Minute)},
	}
	for _, input := range invalid {
		if _, err := NewGrant(input, now); !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("input=%+v err=%v", input, err)
		}
	}
}

func TestGrantCannotBeConsumedTwice(t *testing.T) {
	now := time.Now()
	grant, _ := NewGrant(GrantInput{ObjectKey: "x", OwnerID: 1, Purpose: PurposeAvatar, ContentType: "image/jpeg", Size: 1, ExpiresAt: now.Add(time.Minute)}, now)
	if err := grant.Consume(1, 0, PurposeAvatar, now); err != nil {
		t.Fatal(err)
	}
	if err := grant.Consume(1, 0, PurposeAvatar, now); !errors.Is(err, ErrGrantConsumed) {
		t.Fatalf("err=%v", err)
	}
}
