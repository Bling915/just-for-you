package wish

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type ListWishesQuery struct {
	UserID int64
}

type ListWishes struct {
	relationships ActiveCoupleReader
	store         WishStore
}

func NewListWishes(relationships ActiveCoupleReader, store WishStore) *ListWishes {
	return &ListWishes{relationships: relationships, store: store}
}

func (usecase *ListWishes) Execute(
	ctx context.Context,
	query ListWishesQuery,
) ([]domain.Wish, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	return usecase.store.List(ctx, relationship.ID)
}
