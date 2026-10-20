package application

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	"github.com/nanfang/just-for-you/server/internal/moment/domain"
	"time"
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
	PhotoURLs           map[string]string
}
type Service struct {
	relationships ActiveRelationshipReader
	uploads       UploadAuthorizer
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewService(relationships ActiveRelationshipReader, uploads UploadAuthorizer, moments MomentRepository, photos PhotoStorage, clock appshared.Clock) *Service {
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
	for _, grant := range grants {
		if err := s.photos.VerifyUploaded(ctx, grant.ObjectKey, grant.ContentType, 10<<20); err != nil {
			return domain.Moment{}, err
		}
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
	page, err := s.moments.List(ctx, relationship.ID, cursor, limit)
	if err != nil {
		return TimelinePage{}, err
	}
	page.PhotoURLs, err = s.signReadURLs(ctx, page.Moments)
	return page, err
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
	if err != nil {
		return Home{}, err
	}
	moments := append([]domain.Moment(nil), recent...)
	if latest != nil {
		moments = append(moments, *latest)
	}
	photoURLs, err := s.signReadURLs(ctx, moments)
	return Home{PartnerLatestStatus: latest, Recent: recent, PhotoURLs: photoURLs}, err
}

func (s *Service) signReadURLs(ctx context.Context, moments []domain.Moment) (map[string]string, error) {
	result := make(map[string]string)
	expiresAt := s.clock.Now().Add(10 * time.Minute)
	for _, moment := range moments {
		for _, photo := range moment.Photos() {
			if _, exists := result[photo.ObjectKey()]; exists {
				continue
			}
			url, err := s.photos.SignRead(ctx, photo.ObjectKey(), expiresAt)
			if err != nil {
				return nil, err
			}
			result[photo.ObjectKey()] = url
		}
	}
	return result, nil
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
