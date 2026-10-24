package companionship

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

const nudgeCooldown = 10 * time.Second

type SendNudgeCommand struct {
	UserID int64
	Kind   string
}
type SendNudge struct {
	relationships ActiveCoupleReader
	nudges        NudgeStore
	clock         appshared.Clock
}

func NewSendNudge(r ActiveCoupleReader, store NudgeStore, clock appshared.Clock) *SendNudge {
	return &SendNudge{relationships: r, nudges: store, clock: clock}
}
func (u *SendNudge) Execute(ctx context.Context, command SendNudgeCommand) (domain.Nudge, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Nudge{}, err
	}
	now := u.clock.Now()
	last, err := u.nudges.LastNudgeAt(ctx, relationship.ID, command.UserID)
	if err != nil {
		return domain.Nudge{}, err
	}
	if last != nil && now.Sub(*last) < nudgeCooldown {
		return domain.Nudge{}, domain.ErrNudgeThrottled
	}
	nudge, err := domain.NewNudge(relationship.ID, command.UserID, relationship.UserAID, relationship.UserBID, domain.NudgeKind(command.Kind), now)
	if err != nil {
		return domain.Nudge{}, err
	}
	return u.nudges.SaveNudge(ctx, nudge)
}
