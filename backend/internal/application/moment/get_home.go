package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

type GetHomeQuery struct{ UserID int64 }

// pinnedDisplayLimit 是首页最多展示的置顶纪念日数量。
// 超出部分不丢弃，通过 HasMorePinned 告知前端去「纪念日」页查看。
const pinnedDisplayLimit = 3

type HomeResult struct {
	PartnerLatestStatus *domain.Moment
	Recent              []domain.Moment
	PhotoURLs           map[string]string
	PinnedAnniversaries []AnniversaryHighlight
	HasMorePinned       bool
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
	recent, err := u.moments.Recent(ctx, TimelineFilter{ViewerID: q.UserID, RelationshipID: rel.ID, SpaceType: domain.SpaceShared}, 5)
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
	hasMore := false
	if u.highlights != nil {
		// 多取一个用于判断是否还有更多置顶，避免第 4 个被静默丢弃而用户无感知。
		if values, readErr := u.highlights.ListPinned(ctx, rel.ID, pinnedDisplayLimit+1); readErr == nil {
			if len(values) > pinnedDisplayLimit {
				hasMore = true
				values = values[:pinnedDisplayLimit]
			}
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
	return HomeResult{PartnerLatestStatus: latest, Recent: recent, PhotoURLs: urls, PinnedAnniversaries: highlights, HasMorePinned: hasMore}, nil
}
