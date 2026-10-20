package identity

import (
	"context"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
)

type RefreshSessionCommand struct{ RefreshToken string }

type RefreshSessionUseCase struct{ sessions sessionIssuer }

func NewRefreshSessionUseCase(repository Repository, tokens AccessTokenIssuer, clock appshared.Clock, generator func() (string, error)) *RefreshSessionUseCase {
	return &RefreshSessionUseCase{sessions: sessionIssuer{repository: repository, tokens: tokens, clock: clock, generateRefresh: generator}}
}

func (u *RefreshSessionUseCase) Execute(ctx context.Context, command RefreshSessionCommand) (Session, error) {
	return u.sessions.rotate(ctx, command.RefreshToken)
}
