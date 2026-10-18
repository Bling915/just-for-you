package couple

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct{ DB *sql.DB }

func (r SQLRepository) SaveInvite(c context.Context, i Invite) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO invite_codes(code,creator_user_id,expires_at) VALUES(?,?,?)", i.Code, i.CreatorUserID, i.ExpiresAt)
	return e
}
func (r SQLRepository) Preview(c context.Context, code string) (Invite, error) {
	var x Invite
	var used sql.NullTime
	e := r.DB.QueryRowContext(c, "SELECT code,creator_user_id,expires_at,consumed_at FROM invite_codes WHERE code=?", code).Scan(&x.Code, &x.CreatorUserID, &x.ExpiresAt, &used)
	if errors.Is(e, sql.ErrNoRows) {
		return x, ErrInviteInvalid
	}
	if e != nil {
		return x, e
	}
	if used.Valid {
		return x, ErrInviteUsed
	}
	if !x.ExpiresAt.After(time.Now()) {
		return x, ErrInviteExpired
	}
	return x, nil
}
func (r SQLRepository) Accept(c context.Context, u int64, code string, now time.Time) (Couple, error) {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return Couple{}, e
	}
	defer tx.Rollback()
	var inv Invite
	var used sql.NullTime
	e = tx.QueryRowContext(c, "SELECT code,creator_user_id,expires_at,consumed_at FROM invite_codes WHERE code=? FOR UPDATE", code).Scan(&inv.Code, &inv.CreatorUserID, &inv.ExpiresAt, &used)
	if errors.Is(e, sql.ErrNoRows) {
		return Couple{}, ErrInviteInvalid
	}
	if e != nil {
		return Couple{}, e
	}
	if used.Valid {
		return Couple{}, ErrInviteUsed
	}
	if !inv.ExpiresAt.After(now) {
		return Couple{}, ErrInviteExpired
	}
	if u == inv.CreatorUserID {
		return Couple{}, ErrCannotBindSelf
	}
	var n int
	e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM couples WHERE status='active' AND (user_a_id IN (?,?) OR user_b_id IN (?,?)) FOR UPDATE", u, inv.CreatorUserID, u, inv.CreatorUserID).Scan(&n)
	if e != nil {
		return Couple{}, e
	}
	if n > 0 {
		return Couple{}, ErrAlreadyBound
	}
	a, b := u, inv.CreatorUserID
	if a > b {
		a, b = b, a
	}
	res, e := tx.ExecContext(c, "INSERT INTO couples(user_a_id,user_b_id,status,bound_at) VALUES(?,?,'active',?)", a, b, now)
	if e != nil {
		return Couple{}, ErrAlreadyBound
	}
	id, _ := res.LastInsertId()
	if _, e = tx.ExecContext(c, "UPDATE invite_codes SET consumed_at=?,consumer_user_id=? WHERE code=?", now, u, code); e != nil {
		return Couple{}, e
	}
	return Couple{ID: id, UserAID: a, UserBID: b, Status: "active", BoundAt: now}, tx.Commit()
}
func (r SQLRepository) GetActive(c context.Context, u int64) (Couple, error) {
	var x Couple
	e := r.DB.QueryRowContext(c, "SELECT id,user_a_id,user_b_id,status,bound_at FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?)", u, u).Scan(&x.ID, &x.UserAID, &x.UserBID, &x.Status, &x.BoundAt)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrNoActiveCouple
	}
	return x, e
}
func (r SQLRepository) Unbind(c context.Context, u int64, n time.Time) error {
	res, e := r.DB.ExecContext(c, "UPDATE couples SET status='ended',ended_at=? WHERE status='active' AND (user_a_id=? OR user_b_id=?)", n, u, u)
	if e != nil {
		return e
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNoActiveCouple
	}
	return nil
}
