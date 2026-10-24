package anniversary

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAnniversaryUpdateNormalizesCompleteDraft(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, loc)
	a, err := NewAnniversary(1, 2, " 初见 ", now.AddDate(0, 0, 20), now)
	if err != nil {
		t.Fatal(err)
	}
	err = a.Update(AnniversaryDraft{
		Title:         " 第一次旅行 ",
		Date:          now.AddDate(0, 1, 0),
		CountMode:     CountModeCountdown,
		RepeatsYearly: true,
		Pinned:        true,
		CoverRef:      " couples/1/anniversaries/7/cover.jpg ",
		Message:       " 一起去看更多风景。 ",
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if a.Title() != "第一次旅行" || a.Message() != "一起去看更多风景。" || a.CoverRef() != "couples/1/anniversaries/7/cover.jpg" {
		t.Fatalf("text fields were not normalized: %#v", a)
	}
	if a.CountMode() != CountModeCountdown || !a.RepeatsYearly() || !a.Pinned() {
		t.Fatalf("settings were not applied: %#v", a)
	}
	if !a.UpdatedAt().Equal(now.Add(time.Minute)) {
		t.Fatalf("updatedAt = %v", a.UpdatedAt())
	}
}

func TestAnniversaryUpdateRejectsInvalidFieldsWithoutMutation(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		draft AnniversaryDraft
		want  error
	}{
		{"empty title", AnniversaryDraft{Title: " ", Date: now, CountMode: CountModeElapsed}, ErrInvalidTitle},
		{"long title", AnniversaryDraft{Title: strings.Repeat("纪", 61), Date: now, CountMode: CountModeElapsed}, ErrInvalidTitle},
		{"long message", AnniversaryDraft{Title: "相识", Date: now, CountMode: CountModeElapsed, Message: strings.Repeat("念", 121)}, ErrInvalidMessage},
		{"invalid mode", AnniversaryDraft{Title: "相识", Date: now, CountMode: CountMode("future")}, ErrInvalidCountMode},
		{"zero date", AnniversaryDraft{Title: "相识", CountMode: CountModeElapsed}, ErrInvalidDate},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := NewAnniversary(1, 2, "原始", now, now)
			before := a
			if err := a.Update(tc.draft, now.Add(time.Hour)); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if a != before {
				t.Fatalf("invalid update mutated aggregate: before=%#v after=%#v", before, a)
			}
		})
	}
}

func TestAnniversarySystemItemCannotDelete(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	a := RestoreAnniversary(1, 3, 2, "我们在一起", now, true, CountModeElapsed, true, "", "", "relationship", now, now)
	if a.CanDelete(2) || a.CanDelete(3) {
		t.Fatal("system anniversary must not be deletable")
	}
}

func TestAnniversaryNextOccurrenceHandlesLeapDayAndSameDay(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	created := time.Date(2024, 2, 29, 0, 0, 0, 0, loc)
	a := RestoreAnniversary(1, 3, 2, "相识", created, true, CountModeCountdown, false, "", "", "none", created, created)

	sameDay := time.Date(2025, 2, 28, 23, 59, 0, 0, loc)
	if got := a.NextOccurrence(sameDay); !got.Equal(time.Date(2025, 2, 28, 0, 0, 0, 0, loc)) {
		t.Fatalf("same-day occurrence = %v", got)
	}
	after := time.Date(2025, 3, 1, 0, 0, 0, 0, loc)
	if got := a.NextOccurrence(after); !got.Equal(time.Date(2026, 2, 28, 0, 0, 0, 0, loc)) {
		t.Fatalf("next occurrence = %v", got)
	}
}
