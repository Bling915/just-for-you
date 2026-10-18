package application

import (
	"context"
	"time"

	"github.com/nanfang/just-for-you/server/internal/identity/domain"
)

type ExternalIdentity struct {
	OpenID     string
	SessionKey string
}

type IdentityProvider interface {
	Exchange(context.Context, string) (ExternalIdentity, error)
}

type Repository interface {
	FindOrCreateUser(context.Context, string) (domain.User, error)
	GetUser(context.Context, int64) (domain.User, error)
	UpdateUser(context.Context, domain.User) error
	StoreRefresh(context.Context, domain.RefreshToken) error
	RotateRefresh(context.Context, [32]byte, domain.RefreshToken, time.Time) (int64, error)
	RevokeRefresh(context.Context, [32]byte, time.Time) error
}

type AccessTokenIssuer interface {
	IssueAccess(int64, time.Time) (string, time.Time, error)
}
type AccessTokenVerifier interface{ ParseAccess(string) (int64, error) }
