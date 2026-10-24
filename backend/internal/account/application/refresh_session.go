package application

import (
	"context"

	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type RefreshSessionCommand struct{ RefreshToken string }

type RefreshSession struct{ sessions sessionIssuer }

func NewRefreshSession(repository AccountStore, tokens AccessTokenIssuer, clock appshared.Clock, generator func() (string, error)) *RefreshSession {
	return &RefreshSession{sessions: sessionIssuer{repository: repository, tokens: tokens, clock: clock, generateRefresh: generator}}
}

func (u *RefreshSession) Execute(ctx context.Context, command RefreshSessionCommand) (Session, error) {
	return u.sessions.rotate(ctx, command.RefreshToken)
}
