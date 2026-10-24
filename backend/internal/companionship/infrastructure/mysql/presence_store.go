package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
)

type MySQLPresenceStore struct {
	db *sql.DB
}

func NewMySQLPresenceStore(db *sql.DB) MySQLPresenceStore {
	return MySQLPresenceStore{db: db}
}

func (store MySQLPresenceStore) SavePresence(ctx context.Context, presence domain.Presence) (domain.Presence, error) {
	result, err := store.db.ExecContext(
		ctx,
		"INSERT INTO presences(couple_id,user_id,preset,custom_text,expires_at,updated_at) VALUES(?,?,?,?,?,?)",
		presence.CoupleID(),
		presence.UserID(),
		presence.Preset(),
		presence.CustomText(),
		presence.ExpiresAt(),
		presence.UpdatedAt(),
	)
	if err != nil {
		return domain.Presence{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Presence{}, err
	}
	return domain.RestorePresence(
		id,
		presence.CoupleID(),
		presence.UserID(),
		presence.Preset(),
		presence.CustomText(),
		presence.ExpiresAt(),
		presence.UpdatedAt(),
	), nil
}

func (store MySQLPresenceStore) LatestPresence(
	ctx context.Context,
	coupleID, userID int64,
) (*domain.Presence, error) {
	var (
		id, storedCoupleID, storedUserID int64
		preset                           domain.Preset
		text                             string
		expires                          sql.NullTime
		updated                          time.Time
	)
	err := store.db.QueryRowContext(
		ctx,
		"SELECT id,couple_id,user_id,preset,custom_text,expires_at,updated_at FROM presences WHERE couple_id=? AND user_id=? ORDER BY updated_at DESC,id DESC LIMIT 1",
		coupleID,
		userID,
	).Scan(&id, &storedCoupleID, &storedUserID, &preset, &text, &expires, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var expiry *time.Time
	if expires.Valid {
		expiry = &expires.Time
	}
	presence := domain.RestorePresence(id, storedCoupleID, storedUserID, preset, text, expiry, updated)
	return &presence, nil
}
