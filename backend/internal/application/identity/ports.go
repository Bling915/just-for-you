package identity

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
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
	CreateAvatarGrant(context.Context, AvatarGrant) error
	GetAvatarGrant(context.Context, string) (AvatarGrant, error)
	UpdateUserWithAvatarGrant(context.Context, domain.User, string, time.Time) error
	StoreRefresh(context.Context, domain.RefreshToken) error
	RotateRefresh(context.Context, [32]byte, domain.RefreshToken, time.Time) (int64, error)
	RevokeRefresh(context.Context, [32]byte, time.Time) error
}

type AvatarStorage interface {
	SignUpload(context.Context, string, string, time.Time) (string, error)
	VerifyUploaded(context.Context, string, string, int64) error
	SignRead(context.Context, string, time.Time) (string, error)
}
type Clock interface{ Now() time.Time }

type AccessTokenIssuer interface {
	IssueAccess(int64, time.Time) (string, time.Time, error)
}

type AccessTokenVerifier interface {
	ParseAccess(string) (int64, error)
}
