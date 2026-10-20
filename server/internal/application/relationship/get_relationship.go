package relationship

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

type GetRelationshipQuery struct{ UserID int64 }
type GetRelationshipUseCase struct{ repository Repository }

func NewGetRelationshipUseCase(r Repository) *GetRelationshipUseCase {
	return &GetRelationshipUseCase{r}
}
func (u *GetRelationshipUseCase) Execute(ctx context.Context, q GetRelationshipQuery) (domain.Relationship, error) {
	return u.repository.GetActive(ctx, q.UserID)
}
