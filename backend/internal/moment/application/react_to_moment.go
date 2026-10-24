package application

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type ReactToMomentCommand struct {
	UserID, MomentID int64
	Type             domain.ReactionType
	Text             string
}

type DeleteReactionCommand struct {
	UserID, MomentID, ReactionID int64
}

type ReactToMoment struct {
	relationships ActiveCoupleReader
	moments       momentStore
	clock         appshared.Clock
}

func NewReactToMoment(relationships ActiveCoupleReader, moments momentStore, clock appshared.Clock) *ReactToMoment {
	return &ReactToMoment{relationships: relationships, moments: moments, clock: clock}
}

func (u *ReactToMoment) Execute(ctx context.Context, command ReactToMomentCommand) (domain.Reaction, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Reaction{}, err
	}
	value, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, CoupleID: relationship.ID, MomentID: command.MomentID})
	if err != nil {
		return domain.Reaction{}, err
	}
	if err := (AccessPolicy{}).CanInteract(value, command.UserID, &relationship); err != nil {
		return domain.Reaction{}, err
	}
	reaction, err := domain.NewReaction(command.MomentID, command.UserID, command.Type, command.Text, u.clock.Now())
	if err != nil {
		return domain.Reaction{}, err
	}
	return u.moments.UpsertReaction(ctx, reaction, relationship.ID)
}

func (u *ReactToMoment) Delete(ctx context.Context, command DeleteReactionCommand) error {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return err
	}
	value, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, CoupleID: relationship.ID, MomentID: command.MomentID})
	if err != nil {
		return err
	}
	if err := (AccessPolicy{}).CanRead(value, command.UserID, &relationship); err != nil {
		return err
	}
	return u.moments.DeleteReaction(ctx, command.ReactionID, command.UserID)
}
