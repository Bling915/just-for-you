package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
)

type MySQLNudgeStore struct {
	db *sql.DB
}

func NewMySQLNudgeStore(db *sql.DB) MySQLNudgeStore {
	return MySQLNudgeStore{db: db}
}

func (store MySQLNudgeStore) LastNudgeAt(
	ctx context.Context,
	coupleID, senderID int64,
) (*time.Time, error) {
	var created time.Time
	err := store.db.QueryRowContext(
		ctx,
		"SELECT created_at FROM nudges WHERE couple_id=? AND sender_user_id=? ORDER BY created_at DESC,id DESC LIMIT 1",
		coupleID,
		senderID,
	).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (store MySQLNudgeStore) SaveNudge(ctx context.Context, nudge domain.Nudge) (domain.Nudge, error) {
	result, err := store.db.ExecContext(
		ctx,
		"INSERT INTO nudges(couple_id,sender_user_id,recipient_user_id,kind,created_at) VALUES(?,?,?,?,?)",
		nudge.CoupleID(),
		nudge.SenderID(),
		nudge.RecipientID(),
		nudge.Kind(),
		nudge.CreatedAt(),
	)
	if err != nil {
		return domain.Nudge{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Nudge{}, err
	}
	return domain.RestoreNudge(
		id,
		nudge.CoupleID(),
		nudge.SenderID(),
		nudge.RecipientID(),
		nudge.Kind(),
		nudge.CreatedAt(),
	), nil
}

func (store MySQLNudgeStore) LatestNudge(ctx context.Context, coupleID int64) (*domain.Nudge, error) {
	var (
		id, storedCoupleID, sender, recipient int64
		kind                                  domain.NudgeKind
		created                               time.Time
	)
	err := store.db.QueryRowContext(
		ctx,
		"SELECT id,couple_id,sender_user_id,recipient_user_id,kind,created_at FROM nudges WHERE couple_id=? ORDER BY created_at DESC,id DESC LIMIT 1",
		coupleID,
	).Scan(&id, &storedCoupleID, &sender, &recipient, &kind, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	nudge := domain.RestoreNudge(id, storedCoupleID, sender, recipient, kind, created)
	return &nudge, nil
}
