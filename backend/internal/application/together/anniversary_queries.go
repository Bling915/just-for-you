package together

import (
	"context"
	"sort"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

func (s *Service) ListAnniversaries(ctx context.Context, userID int64) ([]domain.Anniversary, error) {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListAnniversaries(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	hasCouple := false
	for _, item := range items {
		if item.SystemKind() == "relationship" {
			hasCouple = true
			break
		}
	}
	if !hasCouple && !r.BoundAt.IsZero() {
		relationshipItem := domain.RestoreAnniversary(0, r.ID, r.UserAID, "我们在一起", r.BoundAt, true, domain.CountModeElapsed, true, "", "", "relationship", r.BoundAt, r.BoundAt)
		relationshipItem, err = s.repo.SaveAnniversary(ctx, relationshipItem)
		if err != nil {
			return nil, err
		}
		items = append(items, relationshipItem)
	}
	now := s.clock.Now()
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.Pinned() != b.Pinned() {
			return a.Pinned()
		}
		if a.CountMode() != b.CountMode() {
			return a.CountMode() == domain.CountModeCountdown
		}
		if a.CountMode() == domain.CountModeCountdown {
			ai, bj := a.NextOccurrence(now), b.NextOccurrence(now)
			if !ai.Equal(bj) {
				return ai.Before(bj)
			}
		} else if !a.Date().Equal(b.Date()) {
			return a.Date().After(b.Date())
		}
		return a.ID() < b.ID()
	})
	return items, nil
}

func (s *Service) GetAnniversary(ctx context.Context, userID, id int64) (domain.Anniversary, error) {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	return s.repo.FindAnniversary(ctx, r.ID, id)
}
