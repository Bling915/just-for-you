package companionship

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

const nudgeCooldown = 10 * time.Second

type SendNudgeCommand struct {
	UserID int64
	Kind   string
}
type SendNudgeUseCase struct {
	relationships CoupleReader
	repository    Repository
	clock         appshared.Clock
}

func NewSendNudgeUseCase(r CoupleReader, repo Repository, c appshared.Clock) *SendNudgeUseCase {
	return &SendNudgeUseCase{r, repo, c}
}
func (u *SendNudgeUseCase) Execute(ctx context.Context, command SendNudgeCommand) (domain.Nudge, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Nudge{}, err
	}
	now := u.clock.Now()
	last, err := u.repository.LastNudgeAt(ctx, relationship.ID, command.UserID)
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
	return u.repository.SaveNudge(ctx, nudge)
}
