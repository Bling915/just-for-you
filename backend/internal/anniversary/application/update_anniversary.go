package anniversary

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type UpdateAnniversaryCommand struct {
	UserID int64
	ID     int64
	Draft  domain.AnniversaryDraft
}

type UpdateAnniversary struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
	storage       CoverStorage
	clock         Clock
}

func NewUpdateAnniversary(
	relationships ActiveCoupleReader,
	store AnniversaryStore,
	storage CoverStorage,
	clock Clock,
) *UpdateAnniversary {
	return &UpdateAnniversary{
		relationships: relationships,
		store:         store,
		storage:       storage,
		clock:         clock,
	}
}

func (usecase *UpdateAnniversary) Execute(
	ctx context.Context,
	command UpdateAnniversaryCommand,
) (domain.Anniversary, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	item, err := usecase.store.Find(ctx, relationship.ID, command.ID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	oldCover := item.CoverRef()
	now := usecase.clock.Now()
	if err = item.Update(command.Draft, now); err != nil {
		return domain.Anniversary{}, err
	}
	if item.CoverRef() == oldCover || item.CoverRef() == "" {
		return usecase.store.Update(ctx, item)
	}
	if usecase.storage == nil {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	grant, err := usecase.store.GetCoverGrant(ctx, item.CoverRef())
	if err != nil ||
		grant.OwnerID != command.UserID ||
		grant.ResourceID != command.ID ||
		grant.Purpose != "anniversary_cover" ||
		grant.ConsumedAt != nil ||
		!grant.ExpiresAt.After(now) {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	if err = usecase.storage.VerifyUploaded(ctx, grant.ObjectKey, grant.ContentType, 5<<20); err != nil {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	return usecase.store.UpdateWithCoverGrant(ctx, item, grant.ObjectKey, now)
}
