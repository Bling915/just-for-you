package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"strings"
)

type BindCoupleCommand struct {
	UserID int64
	Code   string
}
type BindCouple struct {
	repository CoupleStore
	clock      appshared.Clock
}

func NewBindCouple(r CoupleStore, c appshared.Clock) *BindCouple {
	return &BindCouple{r, c}
}
func (u *BindCouple) Execute(ctx context.Context, c BindCoupleCommand) (domain.Couple, error) {
	return u.repository.AcceptInvitation(ctx, strings.ToUpper(strings.TrimSpace(c.Code)), c.UserID, u.clock.Now())
}
