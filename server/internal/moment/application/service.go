package application

import (
	"context"
	"github.com/nanfang/just-for-you/server/internal/moment/domain"
	platformclock "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type PublishInput struct {
	Text      string
	Status    domain.Status
	PlaceName string
	PhotoKeys []string
}
type Home struct {
	PartnerLatestStatus *domain.Moment
	Recent              []domain.Moment
}
type Service struct {
	relationships ActiveRelationshipReader
	uploads       UploadAuthorizer
	moments       MomentRepository
	photos        PhotoStorage
	clock         platformclock.Clock
}

func NewService(relationships ActiveRelationshipReader, uploads UploadAuthorizer, moments MomentRepository, photos PhotoStorage, clock platformclock.Clock) *Service {
	return &Service{relationships: relationships, uploads: uploads, moments: moments, photos: photos, clock: clock}
}
func (s *Service) RequestPhotoUpload(ctx context.Context, userID int64, input UploadInput) (UploadGrant, error) {
	if input.Size < 1 || input.Size > 10<<20 || !(input.ContentType == "image/jpeg" || input.ContentType == "image/png" || input.ContentType == "image/webp") {
		return UploadGrant{}, domain.ErrInvalidUpload
	}
	now := s.clock.Now()
	grant, err := s.uploads.CreateGrant(ctx, userID, input, now)
	if err != nil {
		return UploadGrant{}, err
	}
	grant.UploadURL, err = s.photos.SignUpload(ctx, grant.ObjectKey, grant.ContentType, grant.ExpiresAt)
	return grant, err
}
func (s *Service) PublishMoment(ctx context.Context, userID int64, input PublishInput) (domain.Moment, error) {
	relationship, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return domain.Moment{}, err
	}
	grants, err := s.uploads.ValidateGrants(ctx, userID, input.PhotoKeys, s.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	moment, err := domain.NewMoment(userID, relationship.ID, input.Text, input.Status, input.PlaceName, input.PhotoKeys, s.clock.Now())
	if err != nil {
		return domain.Moment{}, err
	}
	return s.moments.SaveWithGrants(ctx, moment, grants)
}
func (s *Service) GetTimeline(ctx context.Context, userID int64, cursor string, limit int) (TimelinePage, error) {
	relationship, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return TimelinePage{}, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	return s.moments.List(ctx, relationship.ID, cursor, limit)
}
func (s *Service) GetHome(ctx context.Context, userID int64) (Home, error) {
	relationship, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return Home{}, err
	}
	latest, err := s.moments.LatestStatusByAuthor(ctx, relationship.ID, relationship.PartnerID)
	if err != nil {
		return Home{}, err
	}
	recent, err := s.moments.Recent(ctx, relationship.ID, 5)
	return Home{PartnerLatestStatus: latest, Recent: recent}, err
}
func (s *Service) DeleteMoment(ctx context.Context, userID, momentID int64) error {
	relationship, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return err
	}
	moment, err := s.moments.Find(ctx, relationship.ID, momentID)
	if err != nil {
		return err
	}
	if moment.AuthorID() != userID {
		return domain.ErrNotAuthor
	}
	return s.moments.Delete(ctx, momentID)
}
