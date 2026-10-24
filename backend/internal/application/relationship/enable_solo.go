package relationship

import (
	"context"

	appshared "github.com/nanfang/just-for-you/server/internal/platform/clock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

type EnableSoloTestCommand struct{ UserID int64 }

type EnableSoloTestUseCase struct {
	repository Repository
	clock      appshared.Clock
}

func NewEnableSoloTestUseCase(repository Repository, clock appshared.Clock) *EnableSoloTestUseCase {
	return &EnableSoloTestUseCase{repository: repository, clock: clock}
}

func (u *EnableSoloTestUseCase) Execute(ctx context.Context, command EnableSoloTestCommand) (domain.Relationship, error) {
	return u.repository.CreateSoloRelationship(ctx, command.UserID, u.clock.Now())
}
