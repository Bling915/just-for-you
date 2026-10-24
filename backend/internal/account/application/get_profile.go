package application

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
)

type GetProfileQuery struct{ UserID int64 }
type Profile struct {
	domain.User
	AvatarURL string `json:"avatarUrl"`
}

type GetProfile struct {
	repository AccountStore
	avatars    AvatarStorage
	clock      Clock
}

func NewGetProfile(repository AccountStore, avatars AvatarStorage, clock Clock) *GetProfile {
	return &GetProfile{repository: repository, avatars: avatars, clock: clock}
}

func (u *GetProfile) Execute(ctx context.Context, query GetProfileQuery) (Profile, error) {
	user, err := u.repository.GetUser(ctx, query.UserID)
	if err != nil {
		return Profile{}, err
	}
	url := ""
	if user.AvatarRef() != "" {
		url, err = u.avatars.SignRead(ctx, user.AvatarRef(), u.clock.Now().Add(15*time.Minute))
	}
	return Profile{User: user, AvatarURL: url}, err
}
