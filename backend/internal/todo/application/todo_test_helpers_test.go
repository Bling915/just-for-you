package todo

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type fixedTodoClock struct {
	now time.Time
}

func (clock fixedTodoClock) Now() time.Time {
	return clock.now
}

type memoryTodoStore struct {
	item domain.Todo
}

func (*memoryTodoStore) List(context.Context, int64) ([]domain.Todo, error) {
	return nil, nil
}

func (store *memoryTodoStore) Save(
	_ context.Context,
	item domain.Todo,
) (domain.Todo, error) {
	store.item = item
	return item, nil
}

func (store *memoryTodoStore) Find(
	context.Context,
	int64,
	int64,
) (domain.Todo, error) {
	return store.item, nil
}

func (store *memoryTodoStore) Update(
	_ context.Context,
	item domain.Todo,
) (domain.Todo, error) {
	store.item = item
	return item, nil
}

func (*memoryTodoStore) Delete(context.Context, int64, int64, int64) error {
	return nil
}
