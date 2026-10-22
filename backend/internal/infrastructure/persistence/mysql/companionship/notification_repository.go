package infrastructure

import (
	"context"
	"database/sql"
	application "github.com/nanfang/just-for-you/server/internal/application/companionship"
)

func (r Repository) GetPreferences(ctx context.Context, coupleID, userID int64) (application.NotificationPreferences, error) {
	var v application.NotificationPreferences
	e := r.db.QueryRowContext(ctx, "SELECT moment_enabled,presence_enabled,nudge_enabled FROM notification_preferences WHERE couple_id=? AND user_id=?", coupleID, userID).Scan(&v.MomentEnabled, &v.PresenceEnabled, &v.NudgeEnabled)
	if e == sql.ErrNoRows {
		return v, nil
	}
	return v, e
}
func (r Repository) SavePreferences(ctx context.Context, coupleID, userID int64, v application.NotificationPreferences) (application.NotificationPreferences, error) {
	_, e := r.db.ExecContext(ctx, "INSERT INTO notification_preferences(couple_id,user_id,moment_enabled,presence_enabled,nudge_enabled) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE moment_enabled=VALUES(moment_enabled),presence_enabled=VALUES(presence_enabled),nudge_enabled=VALUES(nudge_enabled)", coupleID, userID, v.MomentEnabled, v.PresenceEnabled, v.NudgeEnabled)
	return v, e
}
