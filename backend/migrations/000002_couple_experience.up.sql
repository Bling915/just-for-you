ALTER TABLE moments
  ADD COLUMN occurred_at DATETIME(6) NULL AFTER place_name,
  ADD COLUMN updated_at DATETIME(6) NULL AFTER created_at,
  ADD COLUMN latitude DECIMAL(10,7) NULL AFTER place_name,
  ADD COLUMN longitude DECIMAL(11,7) NULL AFTER latitude;

UPDATE moments
SET occurred_at = created_at,
    updated_at = created_at
WHERE occurred_at IS NULL OR updated_at IS NULL;

ALTER TABLE moments
  MODIFY COLUMN occurred_at DATETIME(6) NOT NULL,
  MODIFY COLUMN updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  ADD KEY idx_moments_calendar (couple_id, occurred_at DESC, id DESC);

CREATE TABLE presences (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  preset VARCHAR(32) NOT NULL DEFAULT '',
  custom_text VARCHAR(80) NOT NULL DEFAULT '',
  expires_at DATETIME(6) NULL,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_presences_couple_updated (couple_id, updated_at DESC, id DESC),
  KEY idx_presences_user_updated (user_id, updated_at DESC),
  CONSTRAINT fk_presence_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_presence_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE nudges (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  sender_user_id BIGINT UNSIGNED NOT NULL,
  recipient_user_id BIGINT UNSIGNED NOT NULL,
  kind ENUM('miss_you','hug','cheer_up') NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_nudges_couple_created (couple_id, created_at DESC, id DESC),
  KEY idx_nudges_sender_created (sender_user_id, created_at DESC),
  CONSTRAINT fk_nudge_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_nudge_sender FOREIGN KEY (sender_user_id) REFERENCES users(id),
  CONSTRAINT fk_nudge_recipient FOREIGN KEY (recipient_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE notification_preferences (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  moment_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  presence_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  nudge_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_notification_preference_user (couple_id, user_id),
  CONSTRAINT fk_notification_preference_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_notification_preference_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE notification_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  recipient_user_id BIGINT UNSIGNED NOT NULL,
  scene ENUM('moment','presence','nudge') NOT NULL,
  status ENUM('sent','skipped','failed') NOT NULL,
  error_code VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_notification_deliveries_recipient (recipient_user_id, created_at DESC),
  CONSTRAINT fk_notification_delivery_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_notification_delivery_recipient FOREIGN KEY (recipient_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE anniversaries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(120) NOT NULL,
  anniversary_date DATE NOT NULL,
  repeats_yearly BOOLEAN NOT NULL DEFAULT TRUE,
  system_kind ENUM('none','bound') NOT NULL DEFAULT 'none',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_anniversaries_couple_date (couple_id, anniversary_date, id),
  CONSTRAINT fk_anniversary_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_anniversary_creator FOREIGN KEY (creator_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE todos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL,
  completed_by_user_id BIGINT UNSIGNED NULL,
  title VARCHAR(240) NOT NULL,
  due_at DATETIME(6) NULL,
  status ENUM('open','completed') NOT NULL DEFAULT 'open',
  completed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_todos_couple_status (couple_id, status, due_at, id),
  CONSTRAINT fk_todo_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_todo_creator FOREIGN KEY (creator_user_id) REFERENCES users(id),
  CONSTRAINT fk_todo_completed_by FOREIGN KEY (completed_by_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE wishes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL,
  completed_by_user_id BIGINT UNSIGNED NULL,
  linked_moment_id BIGINT UNSIGNED NULL,
  title VARCHAR(240) NOT NULL,
  description VARCHAR(1000) NOT NULL DEFAULT '',
  status ENUM('planned','completed') NOT NULL DEFAULT 'planned',
  completed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_wishes_couple_status (couple_id, status, id),
  CONSTRAINT fk_wish_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_wish_creator FOREIGN KEY (creator_user_id) REFERENCES users(id),
  CONSTRAINT fk_wish_completed_by FOREIGN KEY (completed_by_user_id) REFERENCES users(id),
  CONSTRAINT fk_wish_moment FOREIGN KEY (linked_moment_id) REFERENCES moments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
