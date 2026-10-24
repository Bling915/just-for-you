package todo

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type ActiveCouple struct {
	ID int64
}

type ActiveCoupleReader interface {
	GetActiveCouple(context.Context, int64) (ActiveCouple, error)
}

type Clock interface {
	Now() time.Time
}

type TodoStore interface {
	List(context.Context, int64) ([]domain.Todo, error)
	Save(context.Context, domain.Todo) (domain.Todo, error)
	Find(context.Context, int64, int64) (domain.Todo, error)
	Update(context.Context, domain.Todo) (domain.Todo, error)
	Delete(context.Context, int64, int64, int64) error
}
