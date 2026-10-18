package auth

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct{ DB *sql.DB }

func (r SQLRepository) FindOrCreateUser(ctx context.Context, o string) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE wechat_openid=?", o).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := r.DB.ExecContext(ctx, "INSERT INTO users(wechat_openid) VALUES(?)", o)
	if err != nil {
		return r.FindOrCreateUser(ctx, o)
	}
	return res.LastInsertId()
}
func (r SQLRepository) StoreRefresh(ctx context.Context, x RefreshRecord) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO refresh_tokens(token_hash,user_id,expires_at) VALUES(?,?,?)", x.Hash[:], x.UserID, x.ExpiresAt)
	return e
}
func (r SQLRepository) RotateRefresh(ctx context.Context, h [32]byte, n RefreshRecord, now time.Time) (int64, error) {
	tx, e := r.DB.BeginTx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	var id int64
	var exp time.Time
	var rev sql.NullTime
	e = tx.QueryRowContext(ctx, "SELECT user_id,expires_at,revoked_at FROM refresh_tokens WHERE token_hash=? FOR UPDATE", h[:]).Scan(&id, &exp, &rev)
	if e != nil || rev.Valid || !exp.After(now) {
		return 0, ErrInvalidRefresh
	}
	if _, e = tx.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at=? WHERE token_hash=?", now, h[:]); e != nil {
		return 0, e
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO refresh_tokens(token_hash,user_id,expires_at) VALUES(?,?,?)", n.Hash[:], id, n.ExpiresAt); e != nil {
		return 0, e
	}
	return id, tx.Commit()
}
func (r SQLRepository) RevokeRefresh(ctx context.Context, h [32]byte, now time.Time) error {
	res, e := r.DB.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", now, h[:])
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return ErrInvalidRefresh
	}
	return nil
}
