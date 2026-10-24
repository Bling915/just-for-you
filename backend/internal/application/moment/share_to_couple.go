package moment

import (
	"context"
	"errors"

	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

var ErrCannotShareMoment = errors.New("moment cannot be shared to couple")

type ShareToCoupleCommand struct {
	UserID, MomentID int64
	Message          string
}

type ShareToCoupleUseCase struct {
	relationships ActiveCoupleReader
	moments       MomentRepository
	clock         appshared.Clock
}

func NewShareToCoupleUseCase(relationships ActiveCoupleReader, moments MomentRepository, clock appshared.Clock) *ShareToCoupleUseCase {
	return &ShareToCoupleUseCase{relationships: relationships, moments: moments, clock: clock}
}

func (u *ShareToCoupleUseCase) Execute(ctx context.Context, command ShareToCoupleCommand) (domain.Moment, error) {
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
	if source.SpaceType() != domain.SpacePersonal || source.Kind() != domain.KindStandard {
		return domain.Moment{}, ErrCannotShareMoment
	}
	return u.moments.ShareToCouple(ctx, source.ID(), command.UserID, relationship.ID, command.Message, u.clock.Now())
}
