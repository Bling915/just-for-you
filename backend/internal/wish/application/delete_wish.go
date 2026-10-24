package wish

import "context"

type DeleteWishCommand struct {
	UserID int64
	ID     int64
}

type DeleteWish struct {
	relationships ActiveCoupleReader
	store         WishStore
}

func NewDeleteWish(relationships ActiveCoupleReader, store WishStore) *DeleteWish {
	return &DeleteWish{relationships: relationships, store: store}
}

func (usecase *DeleteWish) Execute(ctx context.Context, command DeleteWishCommand) error {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return err
	}
	return usecase.store.Delete(ctx, relationship.ID, command.ID, command.UserID)
}
