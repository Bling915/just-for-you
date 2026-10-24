package wish

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type fixedWishClock struct {
	now time.Time
}

func (clock fixedWishClock) Now() time.Time {
	return clock.now
}

type memoryWishStore struct {
	item domain.Wish
}

func (*memoryWishStore) List(context.Context, int64) ([]domain.Wish, error) {
	return nil, nil
}

func (store *memoryWishStore) Save(
	_ context.Context,
	item domain.Wish,
) (domain.Wish, error) {
	store.item = item
	return item, nil
}

func (store *memoryWishStore) Find(
	context.Context,
	int64,
	int64,
) (domain.Wish, error) {
	return store.item, nil
}

func (store *memoryWishStore) Update(
	_ context.Context,
	item domain.Wish,
) (domain.Wish, error) {
	store.item = item
	return item, nil
}

func (*memoryWishStore) Delete(context.Context, int64, int64, int64) error {
	return nil
}
