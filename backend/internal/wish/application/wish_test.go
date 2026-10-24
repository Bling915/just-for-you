package wish

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type fakeWishCouple struct{}

func (fakeWishCouple) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return ActiveCouple{ID: 9}, nil
}

type fakeWishMoments struct{ coupleID int64 }

func (f fakeWishMoments) MomentCoupleID(context.Context, int64) (int64, error) {
	return f.coupleID, nil
}

func TestCompleteWishRejectsMomentFromAnotherCouple(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	created, err := domain.NewWish(9, 7, "去看海", "", now)
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryWishStore{item: created}
	usecase := NewCompleteWish(fakeWishCouple{}, fakeWishMoments{coupleID: 10}, store, fixedWishClock{now})
	momentID := int64(88)

	_, err = usecase.Execute(context.Background(), CompleteWishCommand{UserID: 7, ID: 1, MomentID: &momentID})
	if !errors.Is(err, domain.ErrInvalidMoment) {
		t.Fatalf("err=%v", err)
	}
}
