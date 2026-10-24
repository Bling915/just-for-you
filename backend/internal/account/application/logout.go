package application

import (
	"context"
	"crypto/sha256"

	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type LogoutCommand struct{ RefreshToken string }

type Logout struct {
	repository AccountStore
	clock      appshared.Clock
}

func NewLogout(repository AccountStore, clock appshared.Clock) *Logout {
	return &Logout{repository: repository, clock: clock}
}

func (u *Logout) Execute(ctx context.Context, command LogoutCommand) error {
	return u.repository.RevokeRefresh(ctx, sha256.Sum256([]byte(command.RefreshToken)), u.clock.Now())
}
