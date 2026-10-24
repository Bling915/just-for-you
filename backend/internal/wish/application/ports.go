package wish

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
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

type MomentReader interface {
	MomentCoupleID(context.Context, int64) (int64, error)
}

type WishStore interface {
	List(context.Context, int64) ([]domain.Wish, error)
	Save(context.Context, domain.Wish) (domain.Wish, error)
	Find(context.Context, int64, int64) (domain.Wish, error)
	Update(context.Context, domain.Wish) (domain.Wish, error)
	Delete(context.Context, int64, int64, int64) error
}
