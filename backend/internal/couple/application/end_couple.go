package application

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type EndCoupleCommand struct{ UserID int64 }
type EndCouple struct {
	repository CoupleStore
	clock      appshared.Clock
}

func NewEndCouple(r CoupleStore, c appshared.Clock) *EndCouple {
	return &EndCouple{r, c}
}
func (u *EndCouple) Execute(ctx context.Context, c EndCoupleCommand) error {
	return u.repository.End(ctx, c.UserID, u.clock.Now())
}
