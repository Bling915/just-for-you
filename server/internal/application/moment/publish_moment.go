package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

type PublishMomentCommand struct {
	UserID    int64
	Text      string
	Status    domain.Status
	PlaceName string
	PhotoKeys []string
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
	m, err := domain.NewMoment(c.UserID, rel.ID, c.Text, c.Status, c.PlaceName, c.PhotoKeys, u.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	return u.moments.SaveWithGrants(ctx, m, grants)
}
