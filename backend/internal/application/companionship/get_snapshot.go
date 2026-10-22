package companionship

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
)

type GetSnapshotQuery struct{ UserID int64 }
type Snapshot struct {
	PartnerPresence *domain.Presence
	LatestNudge     *domain.Nudge
}
type GetSnapshotUseCase struct {
	relationships RelationshipReader
	repository    Repository
	clock         appshared.Clock
}

func NewGetSnapshotUseCase(r RelationshipReader, repo Repository, c appshared.Clock) *GetSnapshotUseCase {
	return &GetSnapshotUseCase{r, repo, c}
}
func (u *GetSnapshotUseCase) Execute(ctx context.Context, query GetSnapshotQuery) (Snapshot, error) {
	relationship, err := u.relationships.GetActiveRelationship(ctx, query.UserID)
	if err != nil {
		return Snapshot{}, err
	}
	presence, err := u.repository.LatestPresence(ctx, relationship.ID, relationship.PartnerOf(query.UserID))
	if err != nil {
		return Snapshot{}, err
	}
	if presence != nil && !presence.Effective(u.clock.Now()) {
		presence = nil
	}
	nudge, err := u.repository.LatestNudge(ctx, relationship.ID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{PartnerPresence: presence, LatestNudge: nudge}, nil
}
