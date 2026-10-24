package relationship

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type EndRelationshipCommand struct{ UserID int64 }
type EndRelationshipUseCase struct {
	repository Repository
	clock      appshared.Clock
}

func NewEndRelationshipUseCase(r Repository, c appshared.Clock) *EndRelationshipUseCase {
	return &EndRelationshipUseCase{r, c}
}
func (u *EndRelationshipUseCase) Execute(ctx context.Context, c EndRelationshipCommand) error {
	return u.repository.End(ctx, c.UserID, u.clock.Now())
}
