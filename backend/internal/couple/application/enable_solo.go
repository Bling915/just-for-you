package application

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type EnableSoloTestCommand struct{ UserID int64 }

type EnableSoloTest struct {
	repository CoupleStore
	clock      appshared.Clock
}

func NewEnableSoloTest(repository CoupleStore, clock appshared.Clock) *EnableSoloTest {
	return &EnableSoloTest{repository: repository, clock: clock}
}

func (u *EnableSoloTest) Execute(ctx context.Context, command EnableSoloTestCommand) (domain.Couple, error) {
	return u.repository.CreateSoloCouple(ctx, command.UserID, u.clock.Now())
}
