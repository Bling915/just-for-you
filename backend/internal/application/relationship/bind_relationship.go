package relationship

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"strings"
)

type BindRelationshipCommand struct {
	UserID int64
	Code   string
}
type BindRelationshipUseCase struct {
	repository Repository
	clock      appshared.Clock
}

func NewBindRelationshipUseCase(r Repository, c appshared.Clock) *BindRelationshipUseCase {
	return &BindRelationshipUseCase{r, c}
}
func (u *BindRelationshipUseCase) Execute(ctx context.Context, c BindRelationshipCommand) (domain.Relationship, error) {
	return u.repository.AcceptInvitation(ctx, strings.ToUpper(strings.TrimSpace(c.Code)), c.UserID, u.clock.Now())
}
