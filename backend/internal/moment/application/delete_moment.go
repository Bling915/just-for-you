package application

import (
	"context"
	"errors"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type DeleteMomentCommand struct {
	UserID   int64
	MomentID int64
}
type DeleteMoment struct {
	relationships ActiveCoupleReader
	moments       momentStore
}

func NewDeleteMoment(r ActiveCoupleReader, m momentStore) *DeleteMoment {
	return &DeleteMoment{r, m}
}
func (u *DeleteMoment) Execute(ctx context.Context, c DeleteMomentCommand) error {
	var relationship *ActiveCouple
	current, relationshipErr := u.relationships.GetActiveCouple(ctx, c.UserID)
	if relationshipErr == nil {
		relationship = &current
	} else if !errors.Is(relationshipErr, coupledomain.ErrNoActiveCouple) {
		return relationshipErr
	}
	var relationshipID int64
	if relationship != nil {
		relationshipID = relationship.ID
	}
	m, err := u.moments.Find(ctx, MomentLookup{ViewerID: c.UserID, CoupleID: relationshipID, MomentID: c.MomentID})
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
