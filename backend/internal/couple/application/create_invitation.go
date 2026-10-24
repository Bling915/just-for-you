package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"time"
)

type CreateInvitationCommand struct{ UserID int64 }
type CreateInvitation struct {
	repository   CoupleStore
	clock        appshared.Clock
	generateCode func() (string, error)
}

func NewCreateInvitation(r CoupleStore, c appshared.Clock, g func() (string, error)) *CreateInvitation {
	return &CreateInvitation{r, c, g}
}
func (u *CreateInvitation) Execute(ctx context.Context, c CreateInvitationCommand) (domain.Invitation, error) {
	code, err := u.generateCode()
	if err != nil {
		return domain.Invitation{}, err
	}
	invitation, err := domain.NewInvitation(code, c.UserID, u.clock.Now().Add(10*time.Minute))
	if err != nil {
		return domain.Invitation{}, err
	}
	return invitation, u.repository.SaveInvitation(ctx, invitation)
}
