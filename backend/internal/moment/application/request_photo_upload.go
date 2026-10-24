package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type RequestPhotoUploadCommand struct {
	UserID      int64
	ContentType string
	Size        int64
}
type RequestPhotoUpload struct {
	uploads UploadAuthorizer
	photos  PhotoStorage
	clock   appshared.Clock
}

func NewRequestPhotoUpload(u UploadAuthorizer, p PhotoStorage, c appshared.Clock) *RequestPhotoUpload {
	return &RequestPhotoUpload{u, p, c}
}
func (u *RequestPhotoUpload) Execute(ctx context.Context, c RequestPhotoUploadCommand) (UploadGrant, error) {
	input := UploadInput{ContentType: c.ContentType, Size: c.Size}
	if input.Size < 1 || input.Size > 10<<20 || !(input.ContentType == "image/jpeg" || input.ContentType == "image/png" || input.ContentType == "image/webp") {
		return UploadGrant{}, domain.ErrInvalidUpload
	}
	now := u.clock.Now()
	g, err := u.uploads.CreateGrant(ctx, c.UserID, input, now)
	if err != nil {
		return UploadGrant{}, err
	}
	g.UploadURL, err = u.photos.SignUpload(ctx, g.ObjectKey, g.ContentType, g.ExpiresAt)
	return g, err
}
