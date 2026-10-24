package application

import (
	"context"
	"fmt"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
)

type AvatarGrant struct {
	ObjectKey            string
	OwnerID              int64
	ContentType, Purpose string
	Size                 int64
	ExpiresAt            time.Time
	ConsumedAt           *time.Time
}
type UploadGrant struct {
	ObjectKey   string    `json:"objectKey"`
	UploadURL   string    `json:"uploadURL"`
	ContentType string    `json:"contentType"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
type RequestAvatarUploadCommand struct {
	UserID      int64
	ContentType string
	Size        int64
}
type RequestAvatarUpload struct {
	repo    AccountStore
	storage AvatarStorage
	clock   Clock
	key     func() string
}

func NewRequestAvatarUpload(r AccountStore, s AvatarStorage, c Clock, key func() string) *RequestAvatarUpload {
	return &RequestAvatarUpload{r, s, c, key}
}
func (u *RequestAvatarUpload) Execute(ctx context.Context, c RequestAvatarUploadCommand) (UploadGrant, error) {
	if c.Size < 1 || c.Size > 5<<20 || (c.ContentType != "image/jpeg" && c.ContentType != "image/png" && c.ContentType != "image/webp") {
		return UploadGrant{}, domain.ErrInvalidAvatar
	}
	now := u.clock.Now()
	expires := now.Add(10 * time.Minute)
	objectKey := fmt.Sprintf("users/%d/avatars/%s", c.UserID, u.key())
	grant := AvatarGrant{ObjectKey: objectKey, OwnerID: c.UserID, ContentType: c.ContentType, Purpose: "avatar", Size: c.Size, ExpiresAt: expires}
	if err := u.repo.CreateAvatarGrant(ctx, grant); err != nil {
		return UploadGrant{}, err
	}
	url, err := u.storage.SignUpload(ctx, objectKey, c.ContentType, expires)
	if err != nil {
		return UploadGrant{}, err
	}
	return UploadGrant{objectKey, url, c.ContentType, expires}, nil
}
