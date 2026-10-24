package companionship

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type UpdatePresenceCommand struct {
	UserID             int64
	Preset, CustomText string
	ExpiresAt          *time.Time
}
type UpdatePresence struct {
	relationships ActiveCoupleReader
	presence      PresenceStore
	clock         appshared.Clock
}

func NewUpdatePresence(r ActiveCoupleReader, store PresenceStore, clock appshared.Clock) *UpdatePresence {
	return &UpdatePresence{relationships: r, presence: store, clock: clock}
}
func (u *UpdatePresence) Execute(ctx context.Context, command UpdatePresenceCommand) (domain.Presence, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Presence{}, err
	}
	presence, err := domain.NewPresence(relationship.ID, command.UserID, domain.Preset(command.Preset), command.CustomText, command.ExpiresAt, u.clock.Now())
	if err != nil {
		return domain.Presence{}, err
	}
	return u.presence.SavePresence(ctx, presence)
}
