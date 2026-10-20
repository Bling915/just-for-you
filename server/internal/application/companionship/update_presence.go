package companionship

import (
	"context"
	"time"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
)

type UpdatePresenceCommand struct {
	UserID             int64
	Preset, CustomText string
	ExpiresAt          *time.Time
}
type UpdatePresenceUseCase struct {
	relationships RelationshipReader
	repository    Repository
	clock         appshared.Clock
}

func NewUpdatePresenceUseCase(r RelationshipReader, repo Repository, c appshared.Clock) *UpdatePresenceUseCase {
	return &UpdatePresenceUseCase{r, repo, c}
}
func (u *UpdatePresenceUseCase) Execute(ctx context.Context, command UpdatePresenceCommand) (domain.Presence, error) {
	relationship, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return domain.Presence{}, err
	}
	presence, err := domain.NewPresence(relationship.ID, command.UserID, domain.Preset(command.Preset), command.CustomText, command.ExpiresAt, u.clock.Now())
	if err != nil {
		return domain.Presence{}, err
	}
	return u.repository.SavePresence(ctx, presence)
}
