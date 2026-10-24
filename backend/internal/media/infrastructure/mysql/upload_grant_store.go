package mysql

import (
	"context"
	"database/sql"
	"github.com/nanfang/just-for-you/server/internal/media/domain"
	"time"
)

type UploadGrantStore struct{ db *sql.DB }

func NewUploadGrantStore(db *sql.DB) UploadGrantStore { return UploadGrantStore{db} }
func (s UploadGrantStore) Save(ctx context.Context, g domain.Grant) error {
	var resource any
	if g.ResourceID() > 0 {
		resource = g.ResourceID()
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO upload_grants(user_id,object_key,content_type,purpose,resource_id,expires_at) VALUES(?,?,?,?,?,?)", g.OwnerID(), g.ObjectKey(), g.ContentType(), g.Purpose(), resource, g.ExpiresAt())
	return err
}
func (s UploadGrantStore) Consume(ctx context.Context, key string, owner, resource int64, purpose domain.Purpose, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var gotOwner int64
	var gotResource sql.NullInt64
	var content string
	var gotPurpose domain.Purpose
	var expires time.Time
	var consumed sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT user_id,resource_id,content_type,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", key).Scan(&gotOwner, &gotResource, &content, &gotPurpose, &expires, &consumed); err != nil {
		return err
	}
	var consumedAt *time.Time
	if consumed.Valid {
		consumedAt = &consumed.Time
	}
	g := domain.RestoreGrant(domain.GrantInput{ObjectKey: key, OwnerID: gotOwner, ResourceID: gotResource.Int64, Purpose: gotPurpose, ContentType: content, Size: 1, ExpiresAt: expires}, consumedAt)
	if err = g.Consume(owner, resource, purpose, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", now, key); err != nil {
		return err
	}
	return tx.Commit()
}
