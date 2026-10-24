package mysql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	application "github.com/nanfang/just-for-you/server/internal/couple/application"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	"time"
)

type MySQLCoupleStore struct{ db *sql.DB }

func NewMySQLCoupleStore(db *sql.DB) MySQLCoupleStore { return MySQLCoupleStore{db: db} }
func (r MySQLCoupleStore) SaveInvitation(ctx context.Context, i domain.Invitation) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO invite_codes(code,creator_user_id,expires_at) VALUES(?,?,?)", i.Code(), i.CreatorID(), i.ExpiresAt())
	return err
}
func (r MySQLCoupleStore) FindInvitation(ctx context.Context, code string) (domain.Invitation, error) {
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
func (r MySQLCoupleStore) AcceptInvitation(ctx context.Context, code string, userID int64, now time.Time) (domain.Couple, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Couple{}, err
	}
	defer tx.Rollback()
	var creator int64
	var expires time.Time
	var used sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT code,creator_user_id,expires_at,consumed_at FROM invite_codes WHERE code=? FOR UPDATE", code).Scan(new(string), &creator, &expires, &used); errors.Is(err, sql.ErrNoRows) {
		return domain.Couple{}, domain.ErrInvitationInvalid
	}
	if err != nil {
		return domain.Couple{}, err
	}
	var usedAt *time.Time
	if used.Valid {
		usedAt = &used.Time
	}
	invitation := domain.RestoreInvitation(code, creator, expires, usedAt, 0)
	if err = invitation.CanBeAcceptedBy(userID, now); err != nil {
		return domain.Couple{}, err
	}
	// Lock stable user rows in a deterministic order so different invitations
	// cannot concurrently bind the same user into two active relationships.
	first, second := creator, userID
	if first > second {
		first, second = second, first
	}
	rows, err := tx.QueryContext(ctx, "SELECT id FROM users WHERE id IN (?,?) ORDER BY id FOR UPDATE", first, second)
	if err != nil {
		return domain.Couple{}, err
	}
	locked := 0
	for rows.Next() {
		locked++
	}
	if err := rows.Close(); err != nil {
		return domain.Couple{}, err
	}
	if err := rows.Err(); err != nil {
		return domain.Couple{}, err
	}
	if locked != 2 {
		return domain.Couple{}, domain.ErrInvitationInvalid
	}
	var count int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM couples WHERE status='active' AND (user_a_id IN (?,?) OR user_b_id IN (?,?)) FOR UPDATE", userID, creator, userID, creator).Scan(&count); err != nil {
		return domain.Couple{}, err
	}
	if count > 0 {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	relationship := domain.NewCouple(0, creator, userID, now)
	result, err := tx.ExecContext(ctx, "INSERT INTO couples(user_a_id,user_b_id,status,bound_at) VALUES(?,?,'active',?)", relationship.UserAID(), relationship.UserBID(), now)
	if err != nil {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Couple{}, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE invite_codes SET consumed_at=?,consumer_user_id=? WHERE code=?", now, userID, code); err != nil {
		return domain.Couple{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Couple{}, err
	}
	return domain.NewCouple(id, creator, userID, now), nil
}

func (r MySQLCoupleStore) CreateSoloCouple(ctx context.Context, userID int64, now time.Time) (domain.Couple, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Couple{}, err
	}
	defer tx.Rollback()
	var lockedUserID int64
	if err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE id=? FOR UPDATE", userID).Scan(&lockedUserID); err != nil {
		return domain.Couple{}, err
	}
	var active int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", userID, userID).Scan(&active); err != nil {
		return domain.Couple{}, err
	}
	if active > 0 {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	openid := fmt.Sprintf("solo-test-partner:%d", userID)
	result, err := tx.ExecContext(ctx, "INSERT INTO users(wechat_openid,display_name) VALUES(?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),display_name=VALUES(display_name)", openid, "测试搭子")
	if err != nil {
		return domain.Couple{}, err
	}
	partnerID, err := result.LastInsertId()
	if err != nil {
		return domain.Couple{}, err
	}
	relationship := domain.NewCouple(0, userID, partnerID, now)
	result, err = tx.ExecContext(ctx, "INSERT INTO couples(user_a_id,user_b_id,status,bound_at) VALUES(?,?,'active',?)", relationship.UserAID(), relationship.UserBID(), now)
	if err != nil {
		return domain.Couple{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Couple{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Couple{}, err
	}
	return domain.NewCouple(id, userID, partnerID, now), nil
}
func (r MySQLCoupleStore) GetActive(ctx context.Context, userID int64) (domain.Couple, error) {
	var id, a, b int64
	var status string
	var bound time.Time
	var ended sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT id,user_a_id,user_b_id,status,bound_at,ended_at FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?)", userID, userID).Scan(&id, &a, &b, &status, &bound, &ended)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Couple{}, domain.ErrNoActiveCouple
	}
	if err != nil {
		return domain.Couple{}, err
	}
	var endedAt *time.Time
	if ended.Valid {
		endedAt = &ended.Time
	}
	return domain.RestoreCouple(id, a, b, status, bound, endedAt), nil
}
func (r MySQLCoupleStore) GetActiveDetails(ctx context.Context, userID int64) (domain.Couple, application.PartnerProfile, bool, error) {
	var id, a, b, partnerID int64
	var status, displayName, avatarURL string
	var bound time.Time
	var ended sql.NullTime
	var soloTest bool
	err := r.db.QueryRowContext(ctx, `
		SELECT c.id,c.user_a_id,c.user_b_id,c.status,c.bound_at,c.ended_at,
		       partner.id,partner.display_name,partner.avatar_ref,
		       partner.wechat_openid LIKE 'solo-test-partner:%'
		FROM couples c
		JOIN users partner ON partner.id=CASE WHEN c.user_a_id=? THEN c.user_b_id ELSE c.user_a_id END
		WHERE c.status='active' AND (c.user_a_id=? OR c.user_b_id=?)`, userID, userID, userID).
		Scan(&id, &a, &b, &status, &bound, &ended, &partnerID, &displayName, &avatarURL, &soloTest)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Couple{}, application.PartnerProfile{}, false, domain.ErrNoActiveCouple
	}
	if err != nil {
		return domain.Couple{}, application.PartnerProfile{}, false, err
	}
	var endedAt *time.Time
	if ended.Valid {
		endedAt = &ended.Time
	}
	relationship := domain.RestoreCouple(id, a, b, status, bound, endedAt)
	partner := application.PartnerProfile{UserID: partnerID, DisplayName: displayName, AvatarRef: avatarURL}
	return relationship, partner, soloTest, nil
}
func (r MySQLCoupleStore) End(ctx context.Context, userID int64, now time.Time) error {
	result, err := r.db.ExecContext(ctx, "UPDATE couples SET status='ended',ended_at=? WHERE status='active' AND (user_a_id=? OR user_b_id=?)", now, userID, userID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrNoActiveCouple
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
