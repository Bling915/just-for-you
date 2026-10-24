package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	"time"
)

type CoupleStore interface {
	SaveInvitation(context.Context, domain.Invitation) error
	FindInvitation(context.Context, string) (domain.Invitation, error)
	AcceptInvitation(context.Context, string, int64, time.Time) (domain.Couple, error)
	CreateSoloCouple(context.Context, int64, time.Time) (domain.Couple, error)
	GetActive(context.Context, int64) (domain.Couple, error)
	GetActiveDetails(context.Context, int64) (domain.Couple, PartnerProfile, bool, error)
	End(context.Context, int64, time.Time) error
}
type AvatarURLSigner interface {
	SignRead(context.Context, string, time.Time) (string, error)
}
type Clock interface{ Now() time.Time }
