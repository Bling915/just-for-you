package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

type GetResurfacedMemoryQuery struct{ UserID int64 }
type ResurfacedMemoryResult struct {
	Memory    *domain.Moment
	Reason    string
	PhotoURLs map[string]string
}
type GetResurfacedMemoryUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewGetResurfacedMemoryUseCase(r ActiveRelationshipReader, m MomentRepository, p PhotoStorage, c appshared.Clock) *GetResurfacedMemoryUseCase {
	return &GetResurfacedMemoryUseCase{r, m, p, c}
}
func (u *GetResurfacedMemoryUseCase) Execute(ctx context.Context, q GetResurfacedMemoryQuery) (ResurfacedMemoryResult, error) {
	rel, err := u.relationships.GetActiveRelationship(ctx, q.UserID)
	if err != nil {
		return ResurfacedMemoryResult{}, err
	}
	values, err := u.moments.ResurfaceCandidates(ctx, rel.ID, u.clock.Now(), 30)
	if err != nil {
		return ResurfacedMemoryResult{}, err
	}
	var selected *domain.Moment
	reason := ""
	best := -1
	now := u.clock.Now().In(time.FixedZone("CST", 8*3600))
	for i := range values {
		value := values[i]
		if value.SpaceType() != domain.SpaceShared || value.RelationshipID() != rel.ID {
			continue
		}
		at := value.OccurredAt().In(now.Location())
		score, label := 10, "往日的小事"
		if at.Month() == now.Month() && at.Day() == now.Day() && at.Year() < now.Year() {
			score, label = 100, "去年今天"
		} else if at.Year() == now.AddDate(0, -1, 0).Year() && at.Month() == now.AddDate(0, -1, 0).Month() && at.Day() == now.AddDate(0, -1, 0).Day() {
			score, label = 80, "一个月前"
		}
		if score > best {
			copy := value
			selected = &copy
			reason = label
			best = score
		}
	}
	if selected == nil {
		return ResurfacedMemoryResult{}, nil
	}
	urls, err := signReadURLs(ctx, []domain.Moment{*selected}, u.photos, u.clock)
	if err != nil {
		return ResurfacedMemoryResult{}, err
	}
	return ResurfacedMemoryResult{Memory: selected, Reason: reason, PhotoURLs: urls}, nil
}
