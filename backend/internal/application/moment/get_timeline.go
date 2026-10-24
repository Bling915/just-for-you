package moment

import (
	"context"
	"errors"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"time"
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
	if err != nil {
		return TimelinePage{}, err
	}
	ids := make([]int64, 0, len(page.Moments))
	for _, value := range page.Moments {
		ids = append(ids, value.ID())
	}
	page.Reactions, err = u.moments.ListReactions(ctx, ids)
	if err != nil {
		return TimelinePage{}, err
	}
	page.Perspectives, err = u.moments.ListPerspectives(ctx, ids)
	if err != nil {
		return TimelinePage{}, err
	}
	page.PerspectivePhotoURLs, err = signPerspectiveURLs(ctx, page.Perspectives, u.photos, u.clock)
	return page, err
}

func signPerspectiveURLs(ctx context.Context, values map[int64]domain.Perspective, photos PhotoStorage, clock appshared.Clock) (map[string]string, error) {
	result := map[string]string{}
	for _, value := range values {
		for _, photo := range value.Photos() {
			url, err := photos.SignRead(ctx, photo.ObjectKey(), clock.Now().Add(15*time.Minute))
			if err != nil {
				return nil, err
			}
			result[photo.ObjectKey()] = url
		}
	}
	return result, nil
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
