package identity

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type GetProfileQuery struct{ UserID int64 }
type Profile struct {
	domain.User
	AvatarURL string `json:"avatarUrl"`
}

type GetProfileUseCase struct {
	repository Repository
	avatars    AvatarStorage
	clock      Clock
}

func NewGetProfileUseCase(repository Repository, avatars AvatarStorage, clock Clock) *GetProfileUseCase {
	return &GetProfileUseCase{repository: repository, avatars: avatars, clock: clock}
}

func (u *GetProfileUseCase) Execute(ctx context.Context, query GetProfileQuery) (Profile, error) {
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
