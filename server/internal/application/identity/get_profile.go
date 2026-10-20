package identity

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type GetProfileQuery struct{ UserID int64 }

type GetProfileUseCase struct{ repository Repository }

func NewGetProfileUseCase(repository Repository) *GetProfileUseCase {
	return &GetProfileUseCase{repository: repository}
}

func (u *GetProfileUseCase) Execute(ctx context.Context, query GetProfileQuery) (domain.User, error) {
	return u.repository.GetUser(ctx, query.UserID)
}
