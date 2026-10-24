package anniversary

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type GetAnniversaryQuery struct {
	UserID int64
	ID     int64
}

type GetAnniversary struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
}

func NewGetAnniversary(relationships ActiveCoupleReader, store AnniversaryStore) *GetAnniversary {
	return &GetAnniversary{relationships: relationships, store: store}
}

func (usecase *GetAnniversary) Execute(
	ctx context.Context,
	query GetAnniversaryQuery,
) (domain.Anniversary, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return domain.Anniversary{}, err
	}
	return usecase.store.Find(ctx, relationship.ID, query.ID)
}
