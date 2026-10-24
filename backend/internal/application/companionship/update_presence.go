package companionship

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type UpdatePresenceCommand struct {
	UserID             int64
	Preset, CustomText string
	ExpiresAt          *time.Time
}
type UpdatePresenceUseCase struct {
	relationships CoupleReader
	repository    Repository
	clock         appshared.Clock
}

func NewUpdatePresenceUseCase(r CoupleReader, repo Repository, c appshared.Clock) *UpdatePresenceUseCase {
	return &UpdatePresenceUseCase{r, repo, c}
}
func (u *UpdatePresenceUseCase) Execute(ctx context.Context, command UpdatePresenceCommand) (domain.Presence, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Presence{}, err
	}
	presence, err := domain.NewPresence(relationship.ID, command.UserID, domain.Preset(command.Preset), command.CustomText, command.ExpiresAt, u.clock.Now())
	if err != nil {
		return domain.Presence{}, err
	}
	return u.repository.SavePresence(ctx, presence)
}
