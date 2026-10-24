package application

import (
	"context"
	"errors"
	"time"

	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
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
	SpaceType           domain.SpaceType
	Visibility          domain.Visibility
}

type UpdateMoment struct {
	relationships ActiveCoupleReader
	uploads       UploadAuthorizer
	moments       momentStore
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewUpdateMoment(r ActiveCoupleReader, u UploadAuthorizer, m momentStore, p PhotoStorage, c appshared.Clock) *UpdateMoment {
	return &UpdateMoment{r, u, m, p, c}
}

func (u *UpdateMoment) Execute(ctx context.Context, command UpdateMomentCommand) (domain.Moment, error) {
	var relationship *ActiveCouple
	current, relationshipErr := u.relationships.GetActiveCouple(ctx, command.UserID)
	if relationshipErr == nil {
		relationship = &current
	} else if !errors.Is(relationshipErr, coupledomain.ErrNoActiveCouple) {
		return domain.Moment{}, relationshipErr
	}
	var relationshipID int64
	if relationship != nil {
		relationshipID = relationship.ID
	}
	existing, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, CoupleID: relationshipID, MomentID: command.MomentID})
	if err != nil {
		return domain.Moment{}, err
	}
	if existing.SpaceType() == domain.SpaceShared && relationship == nil {
		return domain.Moment{}, relationshipErr
	}
	if err := (AccessPolicy{}).CanMutate(existing, command.UserID, relationship); err != nil {
		return domain.Moment{}, err
	}
	spaceType := command.SpaceType
	if spaceType == "" {
		spaceType = existing.SpaceType()
	}
	visibility := command.Visibility
	if visibility == "" {
		visibility = existing.Visibility()
	}
	newCoupleID := existing.CoupleID()
	if spaceType == domain.SpacePersonal && visibility == domain.VisibilityPrivate {
		newCoupleID = 0
	} else if relationship != nil {
		newCoupleID = relationship.ID
	} else {
		return domain.Moment{}, relationshipErr
	}
	updated, err := existing.Update(command.UserID, domain.UpdateInput{Text: command.Text, Status: command.Status, Place: command.PlaceName, Latitude: command.Latitude, Longitude: command.Longitude, PhotoKeys: command.PhotoKeys, OccurredAt: command.OccurredAt, SpaceType: spaceType, Visibility: visibility, CoupleID: newCoupleID}, u.clock.Now())
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
