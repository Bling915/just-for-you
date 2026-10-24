package todo

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type ListTodosQuery struct {
	UserID int64
}

type ListTodos struct {
	relationships ActiveCoupleReader
	store         TodoStore
}

func NewListTodos(relationships ActiveCoupleReader, store TodoStore) *ListTodos {
	return &ListTodos{relationships: relationships, store: store}
}

func (usecase *ListTodos) Execute(
	ctx context.Context,
	query ListTodosQuery,
) ([]domain.Todo, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	return usecase.store.List(ctx, relationship.ID)
}
