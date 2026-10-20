package relationship

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"time"
)

type Repository interface {
	SaveInvitation(context.Context, domain.Invitation) error
	FindInvitation(context.Context, string) (domain.Invitation, error)
	AcceptInvitation(context.Context, string, int64, time.Time) (domain.Relationship, error)
	GetActive(context.Context, int64) (domain.Relationship, error)
	End(context.Context, int64, time.Time) error
}
