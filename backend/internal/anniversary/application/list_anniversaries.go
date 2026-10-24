package anniversary

import (
	"context"
	"sort"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type ListAnniversariesQuery struct {
	UserID int64
}

type ListAnniversaries struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
	clock         Clock
}

func NewListAnniversaries(
	relationships ActiveCoupleReader,
	store AnniversaryStore,
	clock Clock,
) *ListAnniversaries {
	return &ListAnniversaries{relationships: relationships, store: store, clock: clock}
}

func (usecase *ListAnniversaries) Execute(
	ctx context.Context,
	query ListAnniversariesQuery,
) ([]domain.Anniversary, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	items, err := usecase.store.List(ctx, relationship.ID)
	if err != nil {
		return nil, err
	}
	if !hasRelationshipAnniversary(items) && !relationship.BoundAt.IsZero() {
		item := domain.RestoreAnniversary(
			0,
			relationship.ID,
			relationship.UserAID,
			"我们在一起",
			relationship.BoundAt,
			true,
			domain.CountModeElapsed,
			true,
			"",
			"",
			"relationship",
			relationship.BoundAt,
			relationship.BoundAt,
		)
		item, err = usecase.store.Save(ctx, item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sortAnniversaries(items, usecase.clock.Now())
	return items, nil
}

func hasRelationshipAnniversary(items []domain.Anniversary) bool {
	for _, item := range items {
		if item.SystemKind() == "relationship" {
			return true
		}
	}
	return false
}

func sortAnniversaries(items []domain.Anniversary, now time.Time) {
	sort.SliceStable(items, func(i, j int) bool {
		first, second := items[i], items[j]
		if first.Pinned() != second.Pinned() {
			return first.Pinned()
		}
		if first.CountMode() != second.CountMode() {
			return first.CountMode() == domain.CountModeCountdown
		}
		if first.CountMode() == domain.CountModeCountdown {
			firstDate, secondDate := first.NextOccurrence(now), second.NextOccurrence(now)
			if !firstDate.Equal(secondDate) {
				return firstDate.Before(secondDate)
			}
		} else if !first.Date().Equal(second.Date()) {
			return first.Date().After(second.Date())
		}
		return first.ID() < second.ID()
	})
}
