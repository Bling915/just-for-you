package identity

import (
	"context"
	"strings"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type LoginCommand struct{ Code string }

type LoginUseCase struct {
	provider IdentityProvider
	sessions sessionIssuer
}

func NewLoginUseCase(provider IdentityProvider, repository Repository, tokens AccessTokenIssuer, clock appshared.Clock, generator func() (string, error)) *LoginUseCase {
	return &LoginUseCase{provider: provider, sessions: sessionIssuer{repository: repository, tokens: tokens, clock: clock, generateRefresh: generator}}
}

func (u *LoginUseCase) Execute(ctx context.Context, command LoginCommand) (Session, error) {
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
