package wish

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type CreateWishCommand struct {
	UserID      int64
	Title       string
	Description string
}

type CreateWish struct {
	relationships ActiveCoupleReader
	store         WishStore
	clock         Clock
}

func NewCreateWish(
	relationships ActiveCoupleReader,
	store WishStore,
	clock Clock,
) *CreateWish {
	return &CreateWish{relationships: relationships, store: store, clock: clock}
}

func (usecase *CreateWish) Execute(
	ctx context.Context,
	command CreateWishCommand,
) (domain.Wish, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Wish{}, err
	}
	item, err := domain.NewWish(
		relationship.ID,
		command.UserID,
		command.Title,
		command.Description,
		usecase.clock.Now(),
	)
	if err != nil {
		return domain.Wish{}, err
	}
	return usecase.store.Save(ctx, item)
}
