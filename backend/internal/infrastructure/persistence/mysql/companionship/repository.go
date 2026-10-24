package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	application "github.com/nanfang/just-for-you/server/internal/application/companionship"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) Repository { return Repository{db} }
func (r Repository) GetActiveCouple(ctx context.Context, userID int64) (application.ActiveCouple, error) {
	var result application.ActiveCouple
	err := r.db.QueryRowContext(ctx, "SELECT id,user_a_id,user_b_id FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?)", userID, userID).Scan(&result.ID, &result.UserAID, &result.UserBID)
	if errors.Is(err, sql.ErrNoRows) {
		return application.ActiveCouple{}, coupledomain.ErrNoActiveCouple
	}
	return result, err
}
func (r Repository) SavePresence(ctx context.Context, p domain.Presence) (domain.Presence, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO presences(couple_id,user_id,preset,custom_text,expires_at,updated_at) VALUES(?,?,?,?,?,?)", p.CoupleID(), p.UserID(), p.Preset(), p.CustomText(), p.ExpiresAt(), p.UpdatedAt())
	if err != nil {
		return domain.Presence{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Presence{}, err
	}
	return domain.RestorePresence(id, p.CoupleID(), p.UserID(), p.Preset(), p.CustomText(), p.ExpiresAt(), p.UpdatedAt()), nil
}
func (r Repository) LatestPresence(ctx context.Context, coupleID, userID int64) (*domain.Presence, error) {
	var id, storedCoupleID, storedUserID int64
	var preset domain.Preset
	var text string
	var expires sql.NullTime
	var updated time.Time
	err := r.db.QueryRowContext(ctx, "SELECT id,couple_id,user_id,preset,custom_text,expires_at,updated_at FROM presences WHERE couple_id=? AND user_id=? ORDER BY updated_at DESC,id DESC LIMIT 1", coupleID, userID).Scan(&id, &storedCoupleID, &storedUserID, &preset, &text, &expires, &updated)
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
func (r Repository) LastNudgeAt(ctx context.Context, coupleID, senderID int64) (*time.Time, error) {
	var created time.Time
	err := r.db.QueryRowContext(ctx, "SELECT created_at FROM nudges WHERE couple_id=? AND sender_user_id=? ORDER BY created_at DESC,id DESC LIMIT 1", coupleID, senderID).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &created, nil
}
func (r Repository) SaveNudge(ctx context.Context, n domain.Nudge) (domain.Nudge, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO nudges(couple_id,sender_user_id,recipient_user_id,kind,created_at) VALUES(?,?,?,?,?)", n.CoupleID(), n.SenderID(), n.RecipientID(), n.Kind(), n.CreatedAt())
	if err != nil {
		return domain.Nudge{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Nudge{}, err
	}
	return domain.RestoreNudge(id, n.CoupleID(), n.SenderID(), n.RecipientID(), n.Kind(), n.CreatedAt()), nil
}
func (r Repository) LatestNudge(ctx context.Context, coupleID int64) (*domain.Nudge, error) {
	var id, storedCoupleID, sender, recipient int64
	var kind domain.NudgeKind
	var created time.Time
	err := r.db.QueryRowContext(ctx, "SELECT id,couple_id,sender_user_id,recipient_user_id,kind,created_at FROM nudges WHERE couple_id=? ORDER BY created_at DESC,id DESC LIMIT 1", coupleID).Scan(&id, &storedCoupleID, &sender, &recipient, &kind, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	nudge := domain.RestoreNudge(id, storedCoupleID, sender, recipient, kind, created)
	return &nudge, nil
}
