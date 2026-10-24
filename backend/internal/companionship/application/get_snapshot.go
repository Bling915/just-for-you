package companionship

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type GetSnapshotQuery struct{ UserID int64 }
type Snapshot struct {
	PartnerPresence *domain.Presence
	LatestNudge     *domain.Nudge
}
type GetSnapshot struct {
	relationships ActiveCoupleReader
	presence      PresenceStore
	nudges        NudgeStore
	clock         appshared.Clock
}

func NewGetSnapshot(
	relationships ActiveCoupleReader,
	presence PresenceStore,
	nudges NudgeStore,
	clock appshared.Clock,
) *GetSnapshot {
	return &GetSnapshot{relationships: relationships, presence: presence, nudges: nudges, clock: clock}
}
func (u *GetSnapshot) Execute(ctx context.Context, query GetSnapshotQuery) (Snapshot, error) {
	relationship, err := u.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return Snapshot{}, err
	}
	presence, err := u.presence.LatestPresence(ctx, relationship.ID, relationship.PartnerOf(query.UserID))
	if err != nil {
		return Snapshot{}, err
	}
	if presence != nil && !presence.Effective(u.clock.Now()) {
		presence = nil
	}
	nudge, err := u.nudges.LatestNudge(ctx, relationship.ID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{PartnerPresence: presence, LatestNudge: nudge}, nil
}
