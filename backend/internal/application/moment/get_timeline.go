package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
)

type GetTimelineQuery struct {
	UserID int64
	Cursor string
	Limit  int
}
type GetTimelineUseCase struct {
	relationships ActiveRelationshipReader
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewGetTimelineUseCase(r ActiveRelationshipReader, m MomentRepository, p PhotoStorage, c appshared.Clock) *GetTimelineUseCase {
	return &GetTimelineUseCase{r, m, p, c}
}
func (u *GetTimelineUseCase) Execute(ctx context.Context, q GetTimelineQuery) (TimelinePage, error) {
	rel, err := u.relationships.GetActiveRelationship(ctx, q.UserID)
	if err != nil {
		return TimelinePage{}, err
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	page, err := u.moments.List(ctx, rel.ID, q.Cursor, q.Limit)
	if err != nil {
		return TimelinePage{}, err
	}
	page.PhotoURLs, err = signReadURLs(ctx, page.Moments, u.photos, u.clock)
	return page, err
}
