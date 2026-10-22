package moment

import (
	"context"
	"errors"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

var ErrInvalidSpaceQuery = errors.New("invalid space query")

type GetTimelineQuery struct {
	UserID    int64
	Cursor    string
	Limit     int
	SpaceType domain.SpaceType
	Owner     TimelineOwner
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
	filter, err := u.timelineFilter(ctx, q)
	if err != nil {
		return TimelinePage{}, err
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	page, err := u.moments.List(ctx, filter, q.Cursor, q.Limit)
	if err != nil {
		return TimelinePage{}, err
	}
	page.PhotoURLs, err = signReadURLs(ctx, page.Moments, u.photos, u.clock)
	return page, err
}

func (u *GetTimelineUseCase) timelineFilter(ctx context.Context, q GetTimelineQuery) (TimelineFilter, error) {
	space := q.SpaceType
	if space == "" {
		space = domain.SpaceShared
	}
	if space == domain.SpacePersonal && (q.Owner == "" || q.Owner == TimelineOwnerMe) {
		return TimelineFilter{ViewerID: q.UserID, OwnerID: q.UserID, SpaceType: domain.SpacePersonal}, nil
	}
	if space != domain.SpaceShared && space != domain.SpacePersonal {
		return TimelineFilter{}, ErrInvalidSpaceQuery
	}
	rel, err := u.relationships.GetActiveRelationship(ctx, q.UserID)
	if err != nil {
		return TimelineFilter{}, err
	}
	if space == domain.SpaceShared {
		if q.Owner != "" {
			return TimelineFilter{}, ErrInvalidSpaceQuery
		}
		return TimelineFilter{ViewerID: q.UserID, RelationshipID: rel.ID, SpaceType: domain.SpaceShared}, nil
	}
	if q.Owner != TimelineOwnerPartner {
		return TimelineFilter{}, ErrInvalidSpaceQuery
	}
	if rel.PartnerID <= 0 || rel.PartnerID == q.UserID {
		return TimelineFilter{}, relationshipdomain.ErrNoActiveRelationship
	}
	return TimelineFilter{ViewerID: q.UserID, RelationshipID: rel.ID, OwnerID: rel.PartnerID, SpaceType: domain.SpacePersonal}, nil
}
