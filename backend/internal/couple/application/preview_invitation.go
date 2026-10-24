package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"strings"
)

type PreviewInvitationQuery struct {
	UserID int64
	Code   string
}
type PreviewInvitation struct {
	repository CoupleStore
	clock      appshared.Clock
}

func NewPreviewInvitation(r CoupleStore, c appshared.Clock) *PreviewInvitation {
	return &PreviewInvitation{r, c}
}
func (u *PreviewInvitation) Execute(ctx context.Context, q PreviewInvitationQuery) (domain.Invitation, error) {
	i, err := u.repository.FindInvitation(ctx, strings.ToUpper(strings.TrimSpace(q.Code)))
	if err != nil {
		return domain.Invitation{}, err
	}
	if err = i.CanBeAcceptedBy(q.UserID, u.clock.Now()); err != nil {
		return domain.Invitation{}, err
	}
	return i, nil
}
