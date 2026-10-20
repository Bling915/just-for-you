package identity

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type UpdateProfileCommand struct {
	UserID      int64
	DisplayName string
	AvatarRef   string
}

type UpdateProfileUseCase struct{ repository Repository }

func NewUpdateProfileUseCase(repository Repository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{repository: repository}
}

func (u *UpdateProfileUseCase) Execute(ctx context.Context, command UpdateProfileCommand) (domain.User, error) {
	user, err := u.repository.GetUser(ctx, command.UserID)
	if err != nil {
		return domain.User{}, err
	}
	if err := user.UpdateProfile(command.DisplayName, command.AvatarRef); err != nil {
		return domain.User{}, err
	}
	if err := u.repository.UpdateUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
