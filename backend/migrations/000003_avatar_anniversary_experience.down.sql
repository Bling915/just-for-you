ALTER TABLE upload_grants
  DROP INDEX idx_upload_purpose_resource,
  DROP COLUMN resource_id,
  DROP COLUMN purpose;

ALTER TABLE anniversaries MODIFY COLUMN system_kind ENUM('none','bound','relationship') NOT NULL DEFAULT 'none';
UPDATE anniversaries SET system_kind='bound' WHERE system_kind='relationship';
ALTER TABLE anniversaries MODIFY COLUMN system_kind ENUM('none','bound') NOT NULL DEFAULT 'none';

ALTER TABLE anniversaries
  DROP COLUMN message,
  DROP COLUMN cover_ref,
  DROP COLUMN pinned,
  DROP COLUMN count_mode;
