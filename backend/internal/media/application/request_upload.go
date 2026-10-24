package application

import (
	"context"
	"fmt"
	"github.com/nanfang/just-for-you/server/internal/media/domain"
	"time"
)

type GrantSaver interface {
	Save(context.Context, domain.Grant) error
}
type UploadSigner interface {
	SignUpload(context.Context, string, string, time.Time) (string, error)
}
type RequestUploadCommand struct {
	OwnerID, ResourceID int64
	Purpose             domain.Purpose
	ContentType         string
	Size                int64
}
type UploadGrant struct {
	ObjectKey   string    `json:"objectKey"`
	UploadURL   string    `json:"uploadURL"`
	ContentType string    `json:"contentType"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
type RequestUpload struct {
	store  GrantSaver
	signer UploadSigner
	now    func() time.Time
	key    func() string
}

func NewRequestUpload(s GrantSaver, o UploadSigner, n func() time.Time, k func() string) *RequestUpload {
	return &RequestUpload{s, o, n, k}
}
func (u *RequestUpload) Request(ctx context.Context, c RequestUploadCommand) (UploadGrant, error) {
	now := u.now()
	prefix := ""
	switch c.Purpose {
	case domain.PurposeAvatar:
		prefix = fmt.Sprintf("avatars/%d", c.OwnerID)
	case domain.PurposeMomentPhoto:
		prefix = fmt.Sprintf("moments/%d", c.OwnerID)
	case domain.PurposeAnniversaryCover:
		prefix = fmt.Sprintf("covers/%d/%d", c.OwnerID, c.ResourceID)
	}
	in := domain.GrantInput{ObjectKey: fmt.Sprintf("%s/%s", prefix, u.key()), OwnerID: c.OwnerID, ResourceID: c.ResourceID, Purpose: c.Purpose, ContentType: c.ContentType, Size: c.Size, ExpiresAt: now.Add(10 * time.Minute)}
	g, err := domain.NewGrant(in, now)
	if err != nil {
		return UploadGrant{}, err
	}
	if err = u.store.Save(ctx, g); err != nil {
		return UploadGrant{}, err
	}
	url, err := u.signer.SignUpload(ctx, g.ObjectKey(), g.ContentType(), g.ExpiresAt())
	return UploadGrant{g.ObjectKey(), url, g.ContentType(), g.ExpiresAt()}, err
}
