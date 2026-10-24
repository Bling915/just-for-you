package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	application "github.com/nanfang/just-for-you/server/internal/companionship/application"
)

type MySQLNotificationPreferencesStore struct {
	db *sql.DB
}

func NewMySQLNotificationPreferencesStore(db *sql.DB) MySQLNotificationPreferencesStore {
	return MySQLNotificationPreferencesStore{db: db}
}

func (store MySQLNotificationPreferencesStore) GetPreferences(
	ctx context.Context,
	coupleID, userID int64,
) (application.NotificationPreferences, error) {
	var value application.NotificationPreferences
	err := store.db.QueryRowContext(
		ctx,
		"SELECT moment_enabled,presence_enabled,nudge_enabled,ai_personalization_enabled FROM notification_preferences WHERE couple_id=? AND user_id=?",
		coupleID,
		userID,
	).Scan(
		&value.MomentEnabled,
		&value.PresenceEnabled,
		&value.NudgeEnabled,
		&value.AIPersonalizationEnabled,
	)
	if errors.Is(err, sql.ErrNoRows) {
		value.AIPersonalizationEnabled = true
		return value, nil
	}
	return value, err
}

func (store MySQLNotificationPreferencesStore) SavePreferences(
	ctx context.Context,
	coupleID, userID int64,
	value application.NotificationPreferences,
) (application.NotificationPreferences, error) {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO notification_preferences(couple_id,user_id,moment_enabled,presence_enabled,nudge_enabled,ai_personalization_enabled) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE moment_enabled=VALUES(moment_enabled),presence_enabled=VALUES(presence_enabled),nudge_enabled=VALUES(nudge_enabled),ai_personalization_enabled=VALUES(ai_personalization_enabled)",
		coupleID,
		userID,
		value.MomentEnabled,
		value.PresenceEnabled,
		value.NudgeEnabled,
		value.AIPersonalizationEnabled,
	)
	return value, err
}
