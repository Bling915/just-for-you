package application

import (
	"context"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type apiClock struct{ now time.Time }

func (c apiClock) Now() time.Time { return c.now }

func TestCreateInvitationHasExplicitExecuteContract(t *testing.T) {
	repository := &memoryRepository{active: map[int64]domain.Couple{}}
	useCase := NewCreateInvitation(repository, apiClock{}, func() (string, error) { return "AB3K9Z", nil })
	invitation, err := useCase.Execute(context.Background(), CreateInvitationCommand{UserID: 1})
	if err != nil || invitation.CreatorID() != 1 {
		t.Fatalf("invitation = %+v, error = %v", invitation, err)
	}
}
