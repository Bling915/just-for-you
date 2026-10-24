package moment

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type RequestPhotoUploadCommand struct {
	UserID      int64
	ContentType string
	Size        int64
}
type RequestPhotoUploadUseCase struct {
	uploads UploadAuthorizer
	photos  PhotoStorage
	clock   appshared.Clock
}

func NewRequestPhotoUploadUseCase(u UploadAuthorizer, p PhotoStorage, c appshared.Clock) *RequestPhotoUploadUseCase {
	return &RequestPhotoUploadUseCase{u, p, c}
}
func (u *RequestPhotoUploadUseCase) Execute(ctx context.Context, c RequestPhotoUploadCommand) (UploadGrant, error) {
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
