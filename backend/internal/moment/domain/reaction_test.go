package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewReactionAcceptsPresetTypesAndNormalizesText(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	for _, reactionType := range []ReactionType{ReactionSeen, ReactionHug, ReactionTogether, ReactionCollect} {
		reaction, err := NewReaction(7, 2, reactionType, "", now)
		if err != nil {
			t.Fatalf("type=%q err=%v", reactionType, err)
		}
		if reaction.Type() != reactionType || reaction.Text() != "" {
			t.Fatalf("reaction=%+v", reaction)
		}
	}
	reaction, err := NewReaction(7, 2, ReactionText, "  今天也辛苦啦  ", now)
	if err != nil {
		t.Fatal(err)
	}
	if reaction.Text() != "今天也辛苦啦" {
		t.Fatalf("text=%q", reaction.Text())
	}
}

func TestNewReactionRejectsInvalidOrOversizedText(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		kind ReactionType
		text string
		want error
	}{
		{"unknown type", ReactionType("unknown"), "", ErrInvalidReaction},
		{"empty custom text", ReactionText, "  ", ErrEmptyReactionText},
		{"preset with text", ReactionHug, "不应携带", ErrInvalidReaction},
		{"text over unicode limit", ReactionText, strings.Repeat("好", 121), ErrReactionTextTooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewReaction(7, 2, test.kind, test.text, now)
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v want=%v", err, test.want)
			}
		})
	}
}

func TestReactionIdentitySupportsIdempotentUpsert(t *testing.T) {
	now := time.Now()
	first, _ := NewReaction(7, 2, ReactionHug, "", now)
	second, _ := NewReaction(7, 2, ReactionHug, "", now.Add(time.Minute))
	if first.Identity() != second.Identity() || first.Identity() != "7:2:hug" {
		t.Fatalf("identities %q %q", first.Identity(), second.Identity())
	}
}
