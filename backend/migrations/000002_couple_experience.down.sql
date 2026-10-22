DROP TABLE IF EXISTS wishes;
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS anniversaries;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS nudges;
DROP TABLE IF EXISTS presences;

ALTER TABLE moments
  DROP INDEX idx_moments_calendar,
  DROP COLUMN longitude,
  DROP COLUMN latitude,
  DROP COLUMN updated_at,
  DROP COLUMN occurred_at;
