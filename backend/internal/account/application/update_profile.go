package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
)

type UpdateProfileCommand struct {
	UserID      int64
	DisplayName string
	AvatarRef   string
}

type UpdateProfile struct {
	repository AccountStore
	storage    AvatarStorage
	clock      Clock
}

func NewUpdateProfile(repository AccountStore, storage AvatarStorage, clock Clock) *UpdateProfile {
	return &UpdateProfile{repository: repository, storage: storage, clock: clock}
}

func (u *UpdateProfile) Execute(ctx context.Context, command UpdateProfileCommand) (Profile, error) {
	user, err := u.repository.GetUser(ctx, command.UserID)
	if err != nil {
		return Profile{}, err
	}
	oldAvatarRef := user.AvatarRef()
	if err := user.UpdateProfile(command.DisplayName, command.AvatarRef); err != nil {
		return Profile{}, err
	}
	if command.AvatarRef == "" || command.AvatarRef == oldAvatarRef {
		if err := u.repository.UpdateUser(ctx, user); err != nil {
			return Profile{}, err
		}
	} else {
		grant, err := u.repository.GetAvatarGrant(ctx, command.AvatarRef)
		now := u.clock.Now()
		if err != nil || grant.OwnerID != command.UserID || grant.Purpose != "avatar" || grant.ConsumedAt != nil || !grant.ExpiresAt.After(now) {
			return Profile{}, domain.ErrInvalidAvatar
		}
		prefix := fmt.Sprintf("users/%d/avatars/", command.UserID)
		if !strings.HasPrefix(command.AvatarRef, prefix) {
			return Profile{}, domain.ErrInvalidAvatar
		}
		if err = u.storage.VerifyUploaded(ctx, grant.ObjectKey, grant.ContentType, 5<<20); err != nil {
			return Profile{}, domain.ErrInvalidAvatar
		}
		if err = u.repository.UpdateUserWithAvatarGrant(ctx, user, command.AvatarRef, now); err != nil {
			return Profile{}, err
		}
	}
	url := ""
	if user.AvatarRef() != "" {
		url, err = u.storage.SignRead(ctx, user.AvatarRef(), u.clock.Now().Add(15*time.Minute))
		if err != nil {
			return Profile{}, err
		}
	}
	return Profile{User: user, AvatarURL: url}, nil
}
