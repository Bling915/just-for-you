package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewMomentDetailsAcceptsOnlyValidSpaceVisibilityPairs(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		relationshipID int64
		space          SpaceType
		visibility     Visibility
		wantErr        bool
	}{
		{"shared couple", 9, SpaceShared, VisibilityCouple, false},
		{"personal private", 0, SpacePersonal, VisibilityPrivate, false},
		{"personal partner", 9, SpacePersonal, VisibilityPartner, false},
		{"shared private", 9, SpaceShared, VisibilityPrivate, true},
		{"personal couple", 9, SpacePersonal, VisibilityCouple, true},
		{"shared without relationship", 0, SpaceShared, VisibilityCouple, true},
		{"partner without relationship", 0, SpacePersonal, VisibilityPartner, true},
		{"private tied to relationship", 9, SpacePersonal, VisibilityPrivate, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			moment, err := NewMomentDetails(CreateInput{
				AuthorID: test.relationshipID + 1, CoupleID: test.relationshipID,
				Text: "一件小事", SpaceType: test.space, Visibility: test.visibility,
				OccurredAt: now, CreatedAt: now,
			})
			if test.wantErr {
				if !errors.Is(err, ErrInvalidSpaceVisibility) {
					t.Fatalf("err=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if moment.SpaceType() != test.space || moment.Visibility() != test.visibility {
				t.Fatalf("space=%q visibility=%q", moment.SpaceType(), moment.Visibility())
			}
		})
	}
}

func TestDailyQuestionMomentRequiresSharedQuestionLink(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	questionID := int64(42)
	valid, err := NewMomentDetails(CreateInput{
		AuthorID: 1, CoupleID: 9, Text: "我们今天的回答", SpaceType: SpaceShared,
		Visibility: VisibilityCouple, DailyQuestionID: &questionID,
		OccurredAt: now, CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if valid.DailyQuestionID() == nil || *valid.DailyQuestionID() != questionID {
		t.Fatalf("daily question id=%v", valid.DailyQuestionID())
	}

	invalidQuestionID := int64(0)
	_, err = NewMomentDetails(CreateInput{
		AuthorID: 1, CoupleID: 9, Text: "缺少题目", SpaceType: SpaceShared,
		Visibility: VisibilityCouple, DailyQuestionID: &invalidQuestionID, OccurredAt: now, CreatedAt: now,
	})
	if !errors.Is(err, ErrInvalidMomentKind) {
		t.Fatalf("err=%v", err)
	}
}

func TestSharedCopyKeepsSourceMomentLink(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	sourceID := int64(21)
	moment, err := NewMomentDetails(CreateInput{
		AuthorID: 1, CoupleID: 9, Text: "分享给我们", SpaceType: SpaceShared,
		Visibility: VisibilityCouple, SourceMomentID: &sourceID,
		OccurredAt: now, CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if moment.SourceMomentID() == nil || *moment.SourceMomentID() != sourceID {
		t.Fatalf("source id=%v", moment.SourceMomentID())
	}
}
