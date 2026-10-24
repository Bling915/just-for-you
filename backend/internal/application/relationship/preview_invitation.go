package relationship

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"strings"
)

type PreviewInvitationQuery struct {
	UserID int64
	Code   string
}
type PreviewInvitationUseCase struct {
	repository Repository
	clock      appshared.Clock
}

func NewPreviewInvitationUseCase(r Repository, c appshared.Clock) *PreviewInvitationUseCase {
	return &PreviewInvitationUseCase{r, c}
}
func (u *PreviewInvitationUseCase) Execute(ctx context.Context, q PreviewInvitationQuery) (domain.Invitation, error) {
	i, err := u.repository.FindInvitation(ctx, strings.ToUpper(strings.TrimSpace(q.Code)))
	if err != nil {
		return domain.Invitation{}, err
	}
	if err = i.CanBeAcceptedBy(q.UserID, u.clock.Now()); err != nil {
		return domain.Invitation{}, err
	}
	return i, nil
}
