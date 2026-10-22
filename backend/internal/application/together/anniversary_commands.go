package together

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

func (s *Service) CreateAnniversary(ctx context.Context, userID int64, draft domain.AnniversaryDraft) (domain.Anniversary, error) {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	now := s.clock.Now()
	a, err := domain.NewAnniversary(r.ID, userID, draft.Title, draft.Date, now)
	if err != nil {
		return domain.Anniversary{}, err
	}
	if draft.CountMode == "" {
		draft.CountMode = domain.CountModeCountdown
	}
	if err = a.Update(draft, now); err != nil {
		return domain.Anniversary{}, err
	}
	return s.repo.SaveAnniversary(ctx, a)
}

func (s *Service) UpdateAnniversary(ctx context.Context, userID, id int64, draft domain.AnniversaryDraft) (domain.Anniversary, error) {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	a, err := s.repo.FindAnniversary(ctx, r.ID, id)
	if err != nil {
		return domain.Anniversary{}, err
	}
	oldCover := a.CoverRef()
	now := s.clock.Now()
	if err = a.Update(draft, now); err != nil {
		return domain.Anniversary{}, err
	}
	if a.CoverRef() == oldCover {
		return s.repo.UpdateAnniversary(ctx, a)
	}
	if a.CoverRef() == "" {
		return s.repo.UpdateAnniversary(ctx, a)
	}
	if s.storage == nil {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	g, err := s.repo.GetCoverGrant(ctx, a.CoverRef())
	if err != nil || g.OwnerID != userID || g.ResourceID != id || g.Purpose != "anniversary_cover" || g.ConsumedAt != nil || !g.ExpiresAt.After(now) {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	if err = s.storage.VerifyUploaded(ctx, g.ObjectKey, g.ContentType, 5<<20); err != nil {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	return s.repo.UpdateAnniversaryWithCoverGrant(ctx, a, g.ObjectKey, now)
}

func (s *Service) DeleteAnniversary(ctx context.Context, userID, id int64) error {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return err
	}
	a, err := s.repo.FindAnniversary(ctx, r.ID, id)
	if err != nil {
		return err
	}
	if a.SystemKind() != "none" {
		return domain.ErrSystemAnniversary
	}
	return s.repo.DeleteAnniversary(ctx, r.ID, id, userID)
}
