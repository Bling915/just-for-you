package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

type GetHomeQuery struct{ UserID int64 }
type HomeResult struct {
	PartnerLatestStatus *domain.Moment
	Recent              []domain.Moment
	PhotoURLs           map[string]string
	PinnedAnniversaries []AnniversaryHighlight
}
type GetHomeUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
	highlights    AnniversaryHighlightsReader
}

func NewGetHomeUseCase(r ActiveRelationshipReader, m MomentRepository, p PhotoStorage, c appshared.Clock, readers ...AnniversaryHighlightsReader) *GetHomeUseCase {
	var highlights AnniversaryHighlightsReader
	if len(readers) > 0 {
		highlights = readers[0]
	}
	return &GetHomeUseCase{relationships: r, moments: m, photos: p, clock: c, highlights: highlights}
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
	if err != nil {
		return HomeResult{}, err
	}
	highlights := []AnniversaryHighlight{}
	if u.highlights != nil {
		if values, readErr := u.highlights.ListPinned(ctx, rel.ID, 3); readErr == nil {
			highlights = values
			for i := range highlights {
				if highlights[i].CoverRef != "" {
					if url, signErr := u.photos.SignRead(ctx, highlights[i].CoverRef, u.clock.Now().Add(15*time.Minute)); signErr == nil {
						highlights[i].CoverURL = url
					}
				}
			}
		}
	}
	return HomeResult{PartnerLatestStatus: latest, Recent: recent, PhotoURLs: urls, PinnedAnniversaries: highlights}, nil
}
