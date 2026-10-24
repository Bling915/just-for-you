package anniversary

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type DeleteAnniversaryCommand struct {
	UserID int64
	ID     int64
}

type DeleteAnniversary struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
}

func NewDeleteAnniversary(
	relationships ActiveCoupleReader,
	store AnniversaryStore,
) *DeleteAnniversary {
	return &DeleteAnniversary{relationships: relationships, store: store}
}

func (usecase *DeleteAnniversary) Execute(
	ctx context.Context,
	command DeleteAnniversaryCommand,
) error {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return err
	}
	item, err := usecase.store.Find(ctx, relationship.ID, command.ID)
	if err != nil {
		return err
	}
	if item.SystemKind() != "none" {
		return domain.ErrSystemAnniversary
	}
	return usecase.store.Delete(ctx, relationship.ID, command.ID, command.UserID)
}
