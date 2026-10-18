package infrastructure

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/nanfang/just-for-you/server/internal/relationship/domain"
	"time"
)

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) MySQLRepository { return MySQLRepository{db: db} }
func (r MySQLRepository) SaveInvitation(ctx context.Context, i domain.Invitation) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO invite_codes(code,creator_user_id,expires_at) VALUES(?,?,?)", i.Code(), i.CreatorID(), i.ExpiresAt())
	return err
}
func (r MySQLRepository) FindInvitation(ctx context.Context, code string) (domain.Invitation, error) {
	var creator, consumer int64
	var expires time.Time
	var used sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT creator_user_id,expires_at,consumed_at,COALESCE(consumer_user_id,0) FROM invite_codes WHERE code=?", code).Scan(&creator, &expires, &used, &consumer)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Invitation{}, domain.ErrInvitationInvalid
	}
	if err != nil {
		return domain.Invitation{}, err
	}
	var usedAt *time.Time
	if used.Valid {
		usedAt = &used.Time
	}
	return domain.RestoreInvitation(code, creator, expires, usedAt, consumer), nil
}
func (r MySQLRepository) AcceptInvitation(ctx context.Context, code string, userID int64, now time.Time) (domain.Relationship, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Relationship{}, err
	}
	defer tx.Rollback()
	var creator int64
	var expires time.Time
	var used sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT code,creator_user_id,expires_at,consumed_at FROM invite_codes WHERE code=? FOR UPDATE", code).Scan(new(string), &creator, &expires, &used); errors.Is(err, sql.ErrNoRows) {
		return domain.Relationship{}, domain.ErrInvitationInvalid
	}
	if err != nil {
		return domain.Relationship{}, err
	}
	var usedAt *time.Time
	if used.Valid {
		usedAt = &used.Time
	}
	invitation := domain.RestoreInvitation(code, creator, expires, usedAt, 0)
	if err = invitation.CanBeAcceptedBy(userID, now); err != nil {
		return domain.Relationship{}, err
	}
	var count int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM couples WHERE status='active' AND (user_a_id IN (?,?) OR user_b_id IN (?,?)) FOR UPDATE", userID, creator, userID, creator).Scan(&count); err != nil {
		return domain.Relationship{}, err
	}
	if count > 0 {
		return domain.Relationship{}, domain.ErrAlreadyBound
	}
	relationship := domain.NewRelationship(0, creator, userID, now)
	result, err := tx.ExecContext(ctx, "INSERT INTO couples(user_a_id,user_b_id,status,bound_at) VALUES(?,?,'active',?)", relationship.UserAID(), relationship.UserBID(), now)
	if err != nil {
		return domain.Relationship{}, domain.ErrAlreadyBound
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Relationship{}, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE invite_codes SET consumed_at=?,consumer_user_id=? WHERE code=?", now, userID, code); err != nil {
		return domain.Relationship{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Relationship{}, err
	}
	return domain.NewRelationship(id, creator, userID, now), nil
}
func (r MySQLRepository) GetActive(ctx context.Context, userID int64) (domain.Relationship, error) {
	var id, a, b int64
	var status string
	var bound time.Time
	var ended sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT id,user_a_id,user_b_id,status,bound_at,ended_at FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?)", userID, userID).Scan(&id, &a, &b, &status, &bound, &ended)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Relationship{}, domain.ErrNoActiveRelationship
	}
	if err != nil {
		return domain.Relationship{}, err
	}
	var endedAt *time.Time
	if ended.Valid {
		endedAt = &ended.Time
	}
	return domain.RestoreRelationship(id, a, b, status, bound, endedAt), nil
}
func (r MySQLRepository) End(ctx context.Context, userID int64, now time.Time) error {
	result, err := r.db.ExecContext(ctx, "UPDATE couples SET status='ended',ended_at=? WHERE status='active' AND (user_a_id=? OR user_b_id=?)", now, userID, userID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrNoActiveRelationship
	}
	return nil
}
func GenerateInvitationCode() (string, error) {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	code := make([]byte, 6)
	for i, value := range raw {
		code[i] = alphabet[int(value)%len(alphabet)]
	}
	return string(code), nil
}
