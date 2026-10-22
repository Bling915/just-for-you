package moment

import (
	"context"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

type ReactToMomentCommand struct {
	UserID, MomentID int64
	Type             domain.ReactionType
	Text             string
}

type DeleteReactionCommand struct {
	UserID, MomentID, ReactionID int64
}

type ReactToMomentUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
	clock         appshared.Clock
}

func NewReactToMomentUseCase(relationships ActiveRelationshipReader, moments MomentRepository, clock appshared.Clock) *ReactToMomentUseCase {
	return &ReactToMomentUseCase{relationships: relationships, moments: moments, clock: clock}
}

func (u *ReactToMomentUseCase) Execute(ctx context.Context, command ReactToMomentCommand) (domain.Reaction, error) {
	relationship, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return domain.Reaction{}, err
	}
	value, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, RelationshipID: relationship.ID, MomentID: command.MomentID})
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

func (u *ReactToMomentUseCase) Delete(ctx context.Context, command DeleteReactionCommand) error {
	relationship, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return err
	}
	value, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, RelationshipID: relationship.ID, MomentID: command.MomentID})
	if err != nil {
		return err
	}
	if err := (AccessPolicy{}).CanRead(value, command.UserID, &relationship); err != nil {
		return err
	}
	return u.moments.DeleteReaction(ctx, command.ReactionID, command.UserID)
}
