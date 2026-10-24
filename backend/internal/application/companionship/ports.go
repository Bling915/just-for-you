package companionship

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
)

type ActiveCouple struct{ ID, UserAID, UserBID int64 }

func (r ActiveCouple) PartnerOf(userID int64) int64 {
	if userID == r.UserAID {
		return r.UserBID
	}
	if userID == r.UserBID {
		return r.UserAID
	}
	return 0
}

type CoupleReader interface {
	GetActiveCouple(context.Context, int64) (ActiveCouple, error)
}
type Repository interface {
	SavePresence(context.Context, domain.Presence) (domain.Presence, error)
	LatestPresence(context.Context, int64, int64) (*domain.Presence, error)
	LastNudgeAt(context.Context, int64, int64) (*time.Time, error)
	SaveNudge(context.Context, domain.Nudge) (domain.Nudge, error)
	LatestNudge(context.Context, int64) (*domain.Nudge, error)
}
