package wish

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type CompleteWishCommand struct {
	UserID   int64
	ID       int64
	MomentID *int64
}

type CompleteWish struct {
	relationships ActiveCoupleReader
	moments       MomentReader
	store         WishStore
	clock         Clock
}

func NewCompleteWish(
	relationships ActiveCoupleReader,
	moments MomentReader,
	store WishStore,
	clock Clock,
) *CompleteWish {
	return &CompleteWish{
		relationships: relationships,
		moments:       moments,
		store:         store,
		clock:         clock,
	}
}

func (usecase *CompleteWish) Execute(
	ctx context.Context,
	command CompleteWishCommand,
) (domain.Wish, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Wish{}, err
	}
	item, err := usecase.store.Find(ctx, relationship.ID, command.ID)
	if err != nil {
		return domain.Wish{}, err
	}
	momentCoupleID := relationship.ID
	if command.MomentID != nil {
		momentCoupleID, err = usecase.moments.MomentCoupleID(ctx, *command.MomentID)
		if err != nil {
			return domain.Wish{}, err
		}
	}
	item, err = item.Complete(
		command.UserID,
		command.MomentID,
		momentCoupleID,
		usecase.clock.Now(),
	)
	if err != nil {
		return domain.Wish{}, err
	}
	return usecase.store.Update(ctx, item)
}

type ReopenWishCommand struct {
	UserID int64
	ID     int64
}

type ReopenWish struct {
	relationships ActiveCoupleReader
	store         WishStore
	clock         Clock
}

func NewReopenWish(
	relationships ActiveCoupleReader,
	store WishStore,
	clock Clock,
) *ReopenWish {
	return &ReopenWish{relationships: relationships, store: store, clock: clock}
}

func (usecase *ReopenWish) Execute(
	ctx context.Context,
	command ReopenWishCommand,
) (domain.Wish, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Wish{}, err
	}
	item, err := usecase.store.Find(ctx, relationship.ID, command.ID)
	if err != nil {
		return domain.Wish{}, err
	}
	return usecase.store.Update(ctx, item.Reopen(usecase.clock.Now()))
}
