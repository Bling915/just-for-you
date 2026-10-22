package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

type PublishMomentCommand struct {
	UserID              int64
	Text                string
	Status              domain.Status
	PlaceName           string
	Latitude, Longitude *float64
	PhotoKeys           []string
	OccurredAt          time.Time
}
type PublishMomentUseCase struct {
	relationships ActiveRelationshipReader
	uploads       UploadAuthorizer
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewPublishMomentUseCase(r ActiveRelationshipReader, u UploadAuthorizer, m MomentRepository, p PhotoStorage, c appshared.Clock) *PublishMomentUseCase {
	return &PublishMomentUseCase{r, u, m, p, c}
}
func (u *PublishMomentUseCase) Execute(ctx context.Context, c PublishMomentCommand) (domain.Moment, error) {
	rel, err := u.relationships.GetActiveRelationship(ctx, c.UserID)
	if err != nil {
		return domain.Moment{}, err
	}
	grants, err := u.uploads.ValidateGrants(ctx, c.UserID, c.PhotoKeys, u.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	for _, g := range grants {
		if err := u.photos.VerifyUploaded(ctx, g.ObjectKey, g.ContentType, 10<<20); err != nil {
			return domain.Moment{}, err
		}
	}
	now := u.clock.Now()
	occurredAt := c.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}
	m, err := domain.NewMomentDetails(domain.CreateInput{AuthorID: c.UserID, RelationshipID: rel.ID, Text: c.Text, Status: c.Status, Place: c.PlaceName, Latitude: c.Latitude, Longitude: c.Longitude, PhotoKeys: c.PhotoKeys, OccurredAt: occurredAt, CreatedAt: now, SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, Kind: domain.KindStandard})
	if err != nil {
		return domain.Moment{}, err
	}
	return u.moments.SaveWithGrants(ctx, m, grants)
}
