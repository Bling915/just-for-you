ALTER TABLE moments
  DROP FOREIGN KEY fk_moment_daily_question,
  DROP INDEX uq_moment_daily_question;

DROP TABLE daily_answers;
DROP TABLE daily_questions;
DROP TABLE moment_reactions;
DROP TABLE moment_perspective_photos;
DROP TABLE moment_perspectives;

ALTER TABLE notification_preferences
  DROP COLUMN ai_personalization_enabled;

ALTER TABLE moments
  DROP FOREIGN KEY fk_moment_source,
  DROP CHECK chk_moment_kind_links,
  DROP CHECK chk_moment_space_visibility,
  DROP INDEX idx_moments_source,
  ADD KEY idx_moments_author_rollback (author_user_id);

ALTER TABLE moments
  DROP INDEX idx_moments_personal_timeline;

DELETE FROM moments WHERE space_type='personal' OR kind='daily_question';

ALTER TABLE moments
  DROP COLUMN daily_question_id,
  DROP COLUMN source_moment_id,
  DROP COLUMN kind,
  DROP COLUMN visibility,
  DROP COLUMN space_type,
  MODIFY COLUMN couple_id BIGINT UNSIGNED NOT NULL;
