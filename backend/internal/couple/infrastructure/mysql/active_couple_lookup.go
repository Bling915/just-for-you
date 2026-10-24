package mysql

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type ActiveCoupleRecord struct {
	ID      int64
	UserAID int64
	UserBID int64
}

type ActiveCoupleLookup struct{ store MySQLCoupleStore }

func NewActiveCoupleLookup(dbStore MySQLCoupleStore) ActiveCoupleLookup {
	return ActiveCoupleLookup{store: dbStore}
}

func (l ActiveCoupleLookup) GetActiveCouple(ctx context.Context, userID int64) (ActiveCoupleRecord, error) {
	value, err := l.store.GetActive(ctx, userID)
	if err != nil {
		return ActiveCoupleRecord{}, err
	}
	return activeCoupleRecord(value), nil
}

func activeCoupleRecord(value domain.Couple) ActiveCoupleRecord {
	return ActiveCoupleRecord{ID: value.ID(), UserAID: value.UserAID(), UserBID: value.UserBID()}
}
