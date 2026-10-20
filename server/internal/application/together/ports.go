package together

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type ActiveRelationship struct {
	ID, UserAID, UserBID int64
	BoundAt              time.Time
}
type RelationshipReader interface {
	GetActiveRelationship(context.Context, int64) (ActiveRelationship, error)
}
type Repository interface {
	ListAnniversaries(context.Context, int64) ([]domain.Anniversary, error)
	SaveAnniversary(context.Context, domain.Anniversary) (domain.Anniversary, error)
	DeleteAnniversary(context.Context, int64, int64, int64) error
	ListTodos(context.Context, int64) ([]domain.Todo, error)
	SaveTodo(context.Context, domain.Todo) (domain.Todo, error)
	FindTodo(context.Context, int64, int64) (domain.Todo, error)
	UpdateTodo(context.Context, domain.Todo) (domain.Todo, error)
	DeleteTodo(context.Context, int64, int64, int64) error
	ListWishes(context.Context, int64) ([]domain.Wish, error)
	SaveWish(context.Context, domain.Wish) (domain.Wish, error)
	FindWish(context.Context, int64, int64) (domain.Wish, error)
	UpdateWish(context.Context, domain.Wish) (domain.Wish, error)
	DeleteWish(context.Context, int64, int64, int64) error
	MomentCoupleID(context.Context, int64) (int64, error)
}
type Clock interface{ Now() time.Time }
