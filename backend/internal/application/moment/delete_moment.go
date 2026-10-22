package moment

import (
	"context"
	"errors"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
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
	var relationship *ActiveRelationship
	current, relationshipErr := u.relationships.GetActiveRelationship(ctx, c.UserID)
	if relationshipErr == nil {
		relationship = &current
	} else if !errors.Is(relationshipErr, relationshipdomain.ErrNoActiveRelationship) {
		return relationshipErr
	}
	var relationshipID int64
	if relationship != nil {
		relationshipID = relationship.ID
	}
	m, err := u.moments.Find(ctx, MomentLookup{ViewerID: c.UserID, RelationshipID: relationshipID, MomentID: c.MomentID})
	if err != nil {
		return err
	}
	if m.SpaceType() == "shared" && relationship == nil {
		return relationshipErr
	}
	if err := (AccessPolicy{}).CanMutate(m, c.UserID, relationship); err != nil {
		return err
	}
	return u.moments.Delete(ctx, c.MomentID)
}
