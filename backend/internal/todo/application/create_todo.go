package todo

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type CreateTodoCommand struct {
	UserID int64
	Title  string
	DueAt  *time.Time
}

type CreateTodo struct {
	relationships ActiveCoupleReader
	store         TodoStore
	clock         Clock
}

func NewCreateTodo(
	relationships ActiveCoupleReader,
	store TodoStore,
	clock Clock,
) *CreateTodo {
	return &CreateTodo{relationships: relationships, store: store, clock: clock}
}

func (usecase *CreateTodo) Execute(
	ctx context.Context,
	command CreateTodoCommand,
) (domain.Todo, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.Todo{}, err
	}
	item, err := domain.NewTodo(
		relationship.ID,
		command.UserID,
		command.Title,
		command.DueAt,
		usecase.clock.Now(),
	)
	if err != nil {
		return domain.Todo{}, err
	}
	return usecase.store.Save(ctx, item)
}
