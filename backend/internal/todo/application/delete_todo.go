package todo

import "context"

type DeleteTodoCommand struct {
	UserID int64
	ID     int64
}

type DeleteTodo struct {
	relationships ActiveCoupleReader
	store         TodoStore
}

func NewDeleteTodo(relationships ActiveCoupleReader, store TodoStore) *DeleteTodo {
	return &DeleteTodo{relationships: relationships, store: store}
}

func (usecase *DeleteTodo) Execute(ctx context.Context, command DeleteTodoCommand) error {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return err
	}
	return usecase.store.Delete(ctx, relationship.ID, command.ID, command.UserID)
}
