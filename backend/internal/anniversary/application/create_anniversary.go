package anniversary

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type CreateAnniversaryCommand struct {
	UserID int64
	Draft  domain.AnniversaryDraft
}

type CreateAnniversary struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
	clock         Clock
}

func NewCreateAnniversary(
	relationships ActiveCoupleReader,
	store AnniversaryStore,
	clock Clock,
) *CreateAnniversary {
	return &CreateAnniversary{relationships: relationships, store: store, clock: clock}
}

func (usecase *CreateAnniversary) Execute(
	ctx context.Context,
	command CreateAnniversaryCommand,
) (domain.Anniversary, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	now := usecase.clock.Now()
	item, err := domain.NewAnniversary(
		relationship.ID,
		command.UserID,
		command.Draft.Title,
		command.Draft.Date,
		now,
	)
	if err != nil {
		return domain.Anniversary{}, err
	}
	draft := command.Draft
	if draft.CountMode == "" {
		draft.CountMode = domain.CountModeCountdown
	}
	if err = item.Update(draft, now); err != nil {
		return domain.Anniversary{}, err
	}
	return usecase.store.Save(ctx, item)
}
