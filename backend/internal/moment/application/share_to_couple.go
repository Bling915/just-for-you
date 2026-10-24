package application

import (
	"context"
	"errors"

	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

var ErrCannotShareMoment = errors.New("moment cannot be shared to couple")

type ShareToCoupleCommand struct {
	UserID, MomentID int64
	Message          string
}

type ShareToCouple struct {
	relationships ActiveCoupleReader
	moments       momentStore
	clock         appshared.Clock
}

func NewShareToCouple(relationships ActiveCoupleReader, moments momentStore, clock appshared.Clock) *ShareToCouple {
	return &ShareToCouple{relationships: relationships, moments: moments, clock: clock}
}

func (u *ShareToCouple) Execute(ctx context.Context, command ShareToCoupleCommand) (domain.Moment, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Moment{}, err
	}
	source, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, CoupleID: relationship.ID, MomentID: command.MomentID})
	if err != nil {
		return domain.Moment{}, err
	}
	if source.AuthorID() != command.UserID {
		return domain.Moment{}, domain.ErrNotAuthor
	}
	if source.SpaceType() != domain.SpacePersonal {
		return domain.Moment{}, ErrCannotShareMoment
	}
	return u.moments.ShareToCouple(ctx, source.ID(), command.UserID, relationship.ID, command.Message, u.clock.Now())
}
