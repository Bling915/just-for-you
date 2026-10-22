package moment

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

type DeleteMomentCommand struct {
	UserID   int64
	MomentID int64
}
type DeleteMomentUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
}

func NewDeleteMomentUseCase(r ActiveRelationshipReader, m MomentRepository) *DeleteMomentUseCase {
	return &DeleteMomentUseCase{r, m}
}
func (u *DeleteMomentUseCase) Execute(ctx context.Context, c DeleteMomentCommand) error {
	rel, err := u.relationships.GetActiveRelationship(ctx, c.UserID)
	if err != nil {
		return err
	}
	m, err := u.moments.Find(ctx, MomentLookup{ViewerID: c.UserID, RelationshipID: rel.ID, MomentID: c.MomentID})
	if err != nil {
		return err
	}
	if m.AuthorID() != c.UserID {
		return domain.ErrNotAuthor
	}
	return u.moments.Delete(ctx, c.MomentID)
}
