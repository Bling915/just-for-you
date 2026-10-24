package identity

import (
	"context"
	"crypto/sha256"

	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type LogoutCommand struct{ RefreshToken string }

type LogoutUseCase struct {
	repository Repository
	clock      appshared.Clock
}

func NewLogoutUseCase(repository Repository, clock appshared.Clock) *LogoutUseCase {
	return &LogoutUseCase{repository: repository, clock: clock}
}

func (u *LogoutUseCase) Execute(ctx context.Context, command LogoutCommand) error {
	return u.repository.RevokeRefresh(ctx, sha256.Sum256([]byte(command.RefreshToken)), u.clock.Now())
}
