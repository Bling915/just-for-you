package todo

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type SetTodoCompletedCommand struct {
	UserID    int64
	ID        int64
	Completed bool
}

type SetTodoCompleted struct {
	relationships ActiveCoupleReader
	store         TodoStore
	clock         Clock
}

func NewSetTodoCompleted(
	relationships ActiveCoupleReader,
	store TodoStore,
	clock Clock,
) *SetTodoCompleted {
	return &SetTodoCompleted{relationships: relationships, store: store, clock: clock}
}

func (usecase *SetTodoCompleted) Execute(
	ctx context.Context,
	command SetTodoCompletedCommand,
) (domain.Todo, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Todo{}, err
	}
	item, err := usecase.store.Find(ctx, relationship.ID, command.ID)
	if err != nil {
		return domain.Todo{}, err
	}
	if command.Completed {
		item = item.Complete(command.UserID, usecase.clock.Now())
	} else {
		item = item.Reopen(usecase.clock.Now())
	}
	return usecase.store.Update(ctx, item)
}
