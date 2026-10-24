package anniversary

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type fixedAnniversaryClock struct {
	now time.Time
}

func (clock fixedAnniversaryClock) Now() time.Time {
	return clock.now
}

type memoryAnniversaryStore struct {
	items          []domain.Anniversary
	grant          CoverUploadGrant
	updatedWithKey string
	deleted        bool
}

func newMemoryAnniversaryStore() *memoryAnniversaryStore {
	return &memoryAnniversaryStore{}
}

func (store *memoryAnniversaryStore) List(context.Context, int64) ([]domain.Anniversary, error) {
	return append([]domain.Anniversary(nil), store.items...), nil
}

func (store *memoryAnniversaryStore) Save(
	_ context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	item = domain.RestoreAnniversary(
		1,
		item.CoupleID(),
		item.CreatorID(),
		item.Title(),
		item.Date(),
		item.RepeatsYearly(),
		item.CountMode(),
		item.Pinned(),
		item.CoverRef(),
		item.Message(),
		item.SystemKind(),
		item.CreatedAt(),
		item.UpdatedAt(),
	)
	store.items = append(store.items, item)
	return item, nil
}

func (store *memoryAnniversaryStore) Find(
	context.Context,
	int64,
	int64,
) (domain.Anniversary, error) {
	return store.items[0], nil
}

func (store *memoryAnniversaryStore) Update(
	_ context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	return item, nil
}

func (*memoryAnniversaryStore) CreateCoverGrant(context.Context, CoverUploadGrant) error {
	return nil
}

func (store *memoryAnniversaryStore) GetCoverGrant(context.Context, string) (CoverUploadGrant, error) {
	return store.grant, nil
}

func (store *memoryAnniversaryStore) UpdateWithCoverGrant(
	_ context.Context,
	item domain.Anniversary,
	key string,
	_ time.Time,
) (domain.Anniversary, error) {
	store.updatedWithKey = key
	return item, nil
}

func (store *memoryAnniversaryStore) Delete(context.Context, int64, int64, int64) error {
	store.deleted = true
	return nil
}
