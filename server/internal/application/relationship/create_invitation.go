package relationship

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"time"
)

type CreateInvitationCommand struct{ UserID int64 }
type CreateInvitationUseCase struct {
	repository   Repository
	clock        appshared.Clock
	generateCode func() (string, error)
}

func NewCreateInvitationUseCase(r Repository, c appshared.Clock, g func() (string, error)) *CreateInvitationUseCase {
	return &CreateInvitationUseCase{r, c, g}
}
func (u *CreateInvitationUseCase) Execute(ctx context.Context, c CreateInvitationCommand) (domain.Invitation, error) {
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
