package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

type GetHomeQuery struct{ UserID int64 }
type HomeResult struct {
	PartnerLatestStatus *domain.Moment
	Recent              []domain.Moment
	PhotoURLs           map[string]string
}
type GetHomeUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewGetHomeUseCase(r ActiveRelationshipReader, m MomentRepository, p PhotoStorage, c appshared.Clock) *GetHomeUseCase {
	return &GetHomeUseCase{r, m, p, c}
}
func (u *GetHomeUseCase) Execute(ctx context.Context, q GetHomeQuery) (HomeResult, error) {
	rel, err := u.relationships.GetActiveRelationship(ctx, q.UserID)
	if err != nil {
		return HomeResult{}, err
	}
	latest, err := u.moments.LatestStatusByAuthor(ctx, rel.ID, rel.PartnerID)
	if err != nil {
		return HomeResult{}, err
	}
	recent, err := u.moments.Recent(ctx, rel.ID, 5)
	if err != nil {
		return HomeResult{}, err
	}
	all := append([]domain.Moment(nil), recent...)
	if latest != nil {
		all = append(all, *latest)
	}
	urls, err := signReadURLs(ctx, all, u.photos, u.clock)
	return HomeResult{latest, recent, urls}, err
}
