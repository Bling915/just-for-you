package moment

import (
	"context"
	"time"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

type UpdateMomentCommand struct {
	UserID, MomentID    int64
	Text                string
	Status              domain.Status
	PlaceName           string
	Latitude, Longitude *float64
	PhotoKeys           []string
	OccurredAt          time.Time
	Version             time.Time
}

type UpdateMomentUseCase struct {
	relationships ActiveRelationshipReader
	uploads       UploadAuthorizer
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewUpdateMomentUseCase(r ActiveRelationshipReader, u UploadAuthorizer, m MomentRepository, p PhotoStorage, c appshared.Clock) *UpdateMomentUseCase {
	return &UpdateMomentUseCase{r, u, m, p, c}
}

func (u *UpdateMomentUseCase) Execute(ctx context.Context, command UpdateMomentCommand) (domain.Moment, error) {
	relationship, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return domain.Moment{}, err
	}
	existing, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, RelationshipID: relationship.ID, MomentID: command.MomentID})
	if err != nil {
		return domain.Moment{}, err
	}
	updated, err := existing.Update(command.UserID, domain.UpdateInput{Text: command.Text, Status: command.Status, Place: command.PlaceName, Latitude: command.Latitude, Longitude: command.Longitude, PhotoKeys: command.PhotoKeys, OccurredAt: command.OccurredAt}, u.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	existingKeys := map[string]struct{}{}
	for _, photo := range existing.Photos() {
		existingKeys[photo.ObjectKey()] = struct{}{}
	}
	newKeys := make([]string, 0)
	for _, key := range command.PhotoKeys {
		if _, exists := existingKeys[key]; !exists {
			newKeys = append(newKeys, key)
		}
	}
	grants, err := u.uploads.ValidateGrants(ctx, command.UserID, newKeys, u.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	for _, grant := range grants {
		if err := u.photos.VerifyUploaded(ctx, grant.ObjectKey, grant.ContentType, 10<<20); err != nil {
			return domain.Moment{}, domain.ErrInvalidUpload
		}
	}
	return u.moments.UpdateWithGrants(ctx, updated, command.Version, grants)
}
