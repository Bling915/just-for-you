package anniversary

import (
	"context"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type fakeAnniversaryCouple struct {
	relationship ActiveCouple
}

func TestListAnniversariesSortsPinnedCountdownBeforeElapsed(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	items := []domain.Anniversary{
		domain.RestoreAnniversary(
			1,
			9,
			7,
			"昨天",
			now.AddDate(-1, 0, 0),
			false,
			domain.CountModeElapsed,
			false,
			"",
			"",
			"none",
			now,
			now,
		),
		domain.RestoreAnniversary(
			2,
			9,
			7,
			"下次旅行",
			now.AddDate(0, 2, 0),
			true,
			domain.CountModeCountdown,
			true,
			"",
			"",
			"none",
			now,
			now,
		),
	}
	store := newMemoryAnniversaryStore()
	store.items = items
	usecase := NewListAnniversaries(
		fakeAnniversaryCouple{relationship: ActiveCouple{ID: 9, UserAID: 7}},
		store,
		fixedAnniversaryClock{now},
	)
	result, err := usecase.Execute(context.Background(), ListAnniversariesQuery{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 || result[0].ID() != 2 || result[1].ID() != 1 {
		t.Fatalf("result=%+v", result)
	}
}

func (f fakeAnniversaryCouple) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return f.relationship, nil
}

func TestListAnniversariesKeepsPinnedAndElapsedOrdering(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	store := newMemoryAnniversaryStore()
	usecase := NewListAnniversaries(
		fakeAnniversaryCouple{relationship: ActiveCouple{ID: 9, UserAID: 7, BoundAt: now.AddDate(-2, 0, 0)}},
		store,
		fixedAnniversaryClock{now},
	)

	items, err := usecase.Execute(context.Background(), ListAnniversariesQuery{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].SystemKind() != "relationship" {
		t.Fatalf("items=%+v", items)
	}
}
