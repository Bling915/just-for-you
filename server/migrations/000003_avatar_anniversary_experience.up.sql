ALTER TABLE anniversaries
  ADD COLUMN count_mode ENUM('elapsed','countdown') NOT NULL DEFAULT 'countdown' AFTER repeats_yearly,
  ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT FALSE AFTER count_mode,
  ADD COLUMN cover_ref VARCHAR(512) NOT NULL DEFAULT '' AFTER pinned,
  ADD COLUMN message VARCHAR(120) NOT NULL DEFAULT '' AFTER cover_ref;

ALTER TABLE anniversaries MODIFY COLUMN system_kind ENUM('none','bound','relationship') NOT NULL DEFAULT 'none';
UPDATE anniversaries SET system_kind='relationship' WHERE system_kind='bound';
ALTER TABLE anniversaries MODIFY COLUMN system_kind ENUM('none','relationship') NOT NULL DEFAULT 'none';

ALTER TABLE upload_grants
  ADD COLUMN purpose ENUM('moment_photo','avatar','anniversary_cover') NOT NULL DEFAULT 'moment_photo' AFTER content_type,
  ADD COLUMN resource_id BIGINT UNSIGNED NULL AFTER purpose,
  ADD KEY idx_upload_purpose_resource (purpose, resource_id, expires_at);
