package together

import (
	"context"
	"fmt"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"time"
)

type CoverUploadGrant struct {
	ObjectKey   string     `json:"objectKey"`
	UploadURL   string     `json:"uploadURL"`
	OwnerID     int64      `json:"-"`
	ResourceID  int64      `json:"-"`
	ContentType string     `json:"contentType"`
	Purpose     string     `json:"-"`
	Size        int64      `json:"-"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	ConsumedAt  *time.Time `json:"-"`
}

func (s *Service) RequestCoverUpload(ctx context.Context, userID, id int64, contentType string, size int64) (CoverUploadGrant, error) {
	r, err := s.relationship(ctx, userID)
	if err != nil {
		return CoverUploadGrant{}, err
	}
	if _, err = s.repo.FindAnniversary(ctx, r.ID, id); err != nil {
		return CoverUploadGrant{}, err
	}
	if size < 1 || size > 5<<20 || (contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp") || s.storage == nil {
		return CoverUploadGrant{}, domain.ErrInvalidCover
	}
	now := s.clock.Now()
	g := CoverUploadGrant{ObjectKey: fmt.Sprintf("couples/%d/anniversaries/%d/%s", r.ID, id, s.key()), OwnerID: userID, ResourceID: id, ContentType: contentType, Purpose: "anniversary_cover", Size: size, ExpiresAt: now.Add(10 * time.Minute)}
	if err = s.repo.CreateCoverGrant(ctx, g); err != nil {
		return CoverUploadGrant{}, err
	}
	g.UploadURL, err = s.storage.SignUpload(ctx, g.ObjectKey, g.ContentType, g.ExpiresAt)
	return g, err
}

func (s *Service) CoverURL(ctx context.Context, coverRef string) (string, error) {
	if coverRef == "" || s.storage == nil {
		return "", nil
	}
	return s.storage.SignRead(ctx, coverRef, s.clock.Now().Add(15*time.Minute))
}
