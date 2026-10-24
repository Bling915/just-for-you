package application

import (
	"context"
	"strings"

	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type LoginCommand struct{ Code string }

type Login struct {
	provider IdentityProvider
	sessions sessionIssuer
}

func NewLogin(provider IdentityProvider, repository AccountStore, tokens AccessTokenIssuer, clock appshared.Clock, generator func() (string, error)) *Login {
	return &Login{provider: provider, sessions: sessionIssuer{repository: repository, tokens: tokens, clock: clock, generateRefresh: generator}}
}

func (u *Login) Execute(ctx context.Context, command LoginCommand) (Session, error) {
	if strings.TrimSpace(command.Code) == "" {
		return Session{}, domain.ErrCodeRequired
	}
	external, err := u.provider.Exchange(ctx, command.Code)
	if err != nil {
		return Session{}, err
	}
	user, err := u.sessions.repository.FindOrCreateUser(ctx, external.OpenID)
	if err != nil {
		return Session{}, err
	}
	return u.sessions.create(ctx, user.ID())
}
