package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	application "github.com/nanfang/just-for-you/server/internal/account/application"
	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
)

type MySQLAccountStore struct{ DB *sql.DB }

func (r MySQLAccountStore) FindOrCreateUser(ctx context.Context, openID string) (domain.User, error) {
	user, err := r.findByOpenID(ctx, openID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, err
	}
	result, err := r.DB.ExecContext(ctx, "INSERT INTO users(wechat_openid) VALUES(?)", openID)
	if err != nil {
		return r.findByOpenID(ctx, openID)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}
	return domain.NewUser(id, openID), nil
}
func (r MySQLAccountStore) findByOpenID(ctx context.Context, openID string) (domain.User, error) {
	var id int64
	var name, avatar string
	err := r.DB.QueryRowContext(ctx, "SELECT id,display_name,avatar_ref FROM users WHERE wechat_openid=?", openID).Scan(&id, &name, &avatar)
	return domain.RestoreUser(id, openID, name, avatar), err
}
func (r MySQLAccountStore) GetUser(ctx context.Context, id int64) (domain.User, error) {
	var openID, name, avatar string
	err := r.DB.QueryRowContext(ctx, "SELECT wechat_openid,display_name,avatar_ref FROM users WHERE id=?", id).Scan(&openID, &name, &avatar)
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrUserNotFound
	}
	return domain.RestoreUser(id, openID, name, avatar), err
}

func (r MySQLAccountStore) GetAccount(ctx context.Context, id int64) (domain.User, error) {
	return r.GetUser(ctx, id)
}
func (r MySQLAccountStore) UpdateUser(ctx context.Context, user domain.User) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE users SET display_name=?,avatar_ref=? WHERE id=?", user.DisplayName(), user.AvatarRef(), user.ID())
	return err
}
func (r MySQLAccountStore) CreateAvatarGrant(ctx context.Context, grant application.AvatarGrant) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO upload_grants(user_id,object_key,content_type,purpose,resource_id,expires_at) VALUES(?,?,?,?,NULL,?)", grant.OwnerID, grant.ObjectKey, grant.ContentType, grant.Purpose, grant.ExpiresAt)
	return err
}
func (r MySQLAccountStore) GetAvatarGrant(ctx context.Context, key string) (application.AvatarGrant, error) {
	var g application.AvatarGrant
	var consumed sql.NullTime
	err := r.DB.QueryRowContext(ctx, "SELECT object_key,user_id,content_type,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=?", key).Scan(&g.ObjectKey, &g.OwnerID, &g.ContentType, &g.Purpose, &g.ExpiresAt, &consumed)
	if errors.Is(err, sql.ErrNoRows) {
		return g, domain.ErrInvalidAvatar
	}
	if consumed.Valid {
		g.ConsumedAt = &consumed.Time
	}
	return g, err
}
func (r MySQLAccountStore) UpdateUserWithAvatarGrant(ctx context.Context, user domain.User, key string, now time.Time) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var owner int64
	var purpose string
	var expires time.Time
	var consumed sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT user_id,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", key).Scan(&owner, &purpose, &expires, &consumed); err != nil || owner != user.ID() || purpose != "avatar" || consumed.Valid || !expires.After(now) {
		return domain.ErrInvalidAvatar
	}
	if _, err = tx.ExecContext(ctx, "UPDATE users SET display_name=?,avatar_ref=? WHERE id=?", user.DisplayName(), user.AvatarRef(), user.ID()); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", now, key); err != nil {
		return err
	}
	return tx.Commit()
}
func (r MySQLAccountStore) StoreRefresh(ctx context.Context, token domain.RefreshToken) error {
	hash := token.Hash()
	_, err := r.DB.ExecContext(ctx, "INSERT INTO refresh_tokens(token_hash,user_id,expires_at) VALUES(?,?,?)", hash[:], token.UserID(), token.ExpiresAt())
	return err
}
func (r MySQLAccountStore) RotateRefresh(ctx context.Context, old [32]byte, next domain.RefreshToken, now time.Time) (int64, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var userID int64
	var expires time.Time
	var revoked sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT user_id,expires_at,revoked_at FROM refresh_tokens WHERE token_hash=? FOR UPDATE", old[:]).Scan(&userID, &expires, &revoked); err != nil || revoked.Valid || !expires.After(now) {
		return 0, domain.ErrInvalidRefreshToken
	}
	if _, err = tx.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at=? WHERE token_hash=?", now, old[:]); err != nil {
		return 0, err
	}
	hash := next.Hash()
	if _, err = tx.ExecContext(ctx, "INSERT INTO refresh_tokens(token_hash,user_id,expires_at) VALUES(?,?,?)", hash[:], userID, next.ExpiresAt()); err != nil {
		return 0, err
	}
	return userID, tx.Commit()
}
func (r MySQLAccountStore) RevokeRefresh(ctx context.Context, hash [32]byte, now time.Time) error {
	result, err := r.DB.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", now, hash[:])
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return domain.ErrInvalidRefreshToken
	}
	return nil
}
