CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  wechat_openid VARCHAR(128) NOT NULL,
  display_name VARCHAR(120) NOT NULL DEFAULT '',
  avatar_ref VARCHAR(1024) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_openid (wechat_openid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE couples (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_a_id BIGINT UNSIGNED NOT NULL,
  user_b_id BIGINT UNSIGNED NOT NULL,
  status ENUM('active','ended') NOT NULL DEFAULT 'active',
  bound_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  ended_at DATETIME(6) NULL,
  active_user_a BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE WHEN status='active' THEN user_a_id ELSE NULL END
  ) STORED,
  active_user_b BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE WHEN status='active' THEN user_b_id ELSE NULL END
  ) STORED,
  PRIMARY KEY (id),
  CONSTRAINT chk_couple_order CHECK (user_a_id < user_b_id),
  CONSTRAINT fk_couple_a FOREIGN KEY (user_a_id) REFERENCES users(id),
  CONSTRAINT fk_couple_b FOREIGN KEY (user_b_id) REFERENCES users(id),
  UNIQUE KEY uq_active_user_a (active_user_a),
  UNIQUE KEY uq_active_user_b (active_user_b),
  KEY idx_couple_a_status (user_a_id, status),
  KEY idx_couple_b_status (user_b_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE invite_codes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code CHAR(6) NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  consumed_at DATETIME(6) NULL,
  consumer_user_id BIGINT UNSIGNED NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_invite_code (code),
  CONSTRAINT fk_invite_creator FOREIGN KEY (creator_user_id) REFERENCES users(id),
  CONSTRAINT fk_invite_consumer FOREIGN KEY (consumer_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  token_hash BINARY(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  revoked_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_refresh_hash (token_hash),
  KEY idx_refresh_user (user_id),
  CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE upload_grants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  object_key VARCHAR(512) NOT NULL,
  content_type VARCHAR(64) NOT NULL,
  purpose ENUM('moment_photo','avatar','anniversary_cover') NOT NULL DEFAULT 'moment_photo',
  resource_id BIGINT UNSIGNED NULL,
  expires_at DATETIME(6) NOT NULL,
  consumed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_upload_object (object_key),
  KEY idx_upload_user (user_id, expires_at),
  KEY idx_upload_purpose_resource (purpose, resource_id, expires_at),
  CONSTRAINT fk_upload_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE daily_questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  question_date DATE NOT NULL,
  question_text VARCHAR(240) NOT NULL,
  source ENUM('ai','fallback') NOT NULL,
  context_version VARCHAR(32) NOT NULL DEFAULT '',
  unlocked_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_daily_question_couple_date (couple_id, question_date),
  KEY idx_daily_questions_couple_created (couple_id, created_at DESC, id DESC),
  CONSTRAINT fk_daily_question_couple FOREIGN KEY (couple_id) REFERENCES couples(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NULL,
  author_user_id BIGINT UNSIGNED NOT NULL,
  space_type ENUM('shared','personal') NOT NULL,
  visibility ENUM('couple','private','partner') NOT NULL,
  source_moment_id BIGINT UNSIGNED NULL,
  daily_question_id BIGINT UNSIGNED NULL,
  body VARCHAR(2000) NOT NULL DEFAULT '',
  status VARCHAR(80) NOT NULL DEFAULT '',
  place_name VARCHAR(320) NOT NULL DEFAULT '',
  latitude DECIMAL(10,7) NULL,
  longitude DECIMAL(11,7) NULL,
  occurred_at DATETIME(6) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  KEY idx_moments_timeline (couple_id, created_at DESC, id DESC),
  KEY idx_moments_calendar (couple_id, occurred_at DESC, id DESC),
  KEY idx_moments_personal_timeline (author_user_id, space_type, visibility, occurred_at DESC, id DESC),
  KEY idx_moments_source (source_moment_id),
  UNIQUE KEY uq_moment_daily_question (daily_question_id),
  CONSTRAINT chk_moment_space_visibility CHECK (
    (space_type='shared' AND visibility='couple' AND couple_id IS NOT NULL)
    OR (space_type='personal' AND visibility IN ('private','partner'))
  ),
  CONSTRAINT chk_moment_daily_question CHECK (
    daily_question_id IS NULL
    OR (space_type='shared' AND visibility='couple')
  ),
  CONSTRAINT fk_moment_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_moment_author FOREIGN KEY (author_user_id) REFERENCES users(id),
  CONSTRAINT fk_moment_source FOREIGN KEY (source_moment_id) REFERENCES moments(id) ON DELETE SET NULL,
  CONSTRAINT fk_moment_daily_question FOREIGN KEY (daily_question_id) REFERENCES daily_questions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moment_photos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  moment_id BIGINT UNSIGNED NOT NULL,
  object_key VARCHAR(1024) NOT NULL,
  sort_order TINYINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_moment_photo_order (moment_id, sort_order),
  CONSTRAINT chk_photo_order CHECK (sort_order BETWEEN 1 AND 9),
  CONSTRAINT fk_photo_moment FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE daily_answers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  daily_question_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  answer_text VARCHAR(500) NOT NULL,
  submitted_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_daily_answer_question_user (daily_question_id, user_id),
  KEY idx_daily_answers_user_submitted (user_id, submitted_at DESC),
  CONSTRAINT fk_daily_answer_question FOREIGN KEY (daily_question_id) REFERENCES daily_questions(id) ON DELETE CASCADE,
  CONSTRAINT fk_daily_answer_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
  ai_personalization_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_notification_preference_user (couple_id, user_id),
  CONSTRAINT fk_notification_preference_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_notification_preference_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE anniversaries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  couple_id BIGINT UNSIGNED NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(120) NOT NULL,
  anniversary_date DATE NOT NULL,
  repeats_yearly BOOLEAN NOT NULL DEFAULT TRUE,
  count_mode ENUM('elapsed','countdown') NOT NULL DEFAULT 'countdown',
  pinned BOOLEAN NOT NULL DEFAULT FALSE,
  cover_ref VARCHAR(512) NOT NULL DEFAULT '',
  message VARCHAR(120) NOT NULL DEFAULT '',
  system_kind ENUM('none','relationship') NOT NULL DEFAULT 'none',
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

CREATE TABLE moment_perspectives (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  moment_id BIGINT UNSIGNED NOT NULL,
  author_user_id BIGINT UNSIGNED NOT NULL,
  body VARCHAR(2000) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_moment_perspective_author (moment_id, author_user_id),
  KEY idx_perspectives_author_updated (author_user_id, updated_at DESC),
  CONSTRAINT fk_perspective_moment FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE,
  CONSTRAINT fk_perspective_author FOREIGN KEY (author_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moment_perspective_photos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  perspective_id BIGINT UNSIGNED NOT NULL,
  object_key VARCHAR(1024) NOT NULL,
  sort_order TINYINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_perspective_photo_order (perspective_id, sort_order),
  CONSTRAINT chk_perspective_photo_order CHECK (sort_order BETWEEN 1 AND 3),
  CONSTRAINT fk_perspective_photo FOREIGN KEY (perspective_id) REFERENCES moment_perspectives(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moment_reactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  moment_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  reaction_type ENUM('seen','hug','together','collect','text') NOT NULL,
  text_content VARCHAR(480) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_moment_reaction_type (moment_id, user_id, reaction_type),
  KEY idx_reactions_moment_created (moment_id, created_at, id),
  CONSTRAINT fk_reaction_moment FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE,
  CONSTRAINT fk_reaction_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT chk_reaction_text CHECK (
    (reaction_type='text' AND CHAR_LENGTH(text_content) BETWEEN 1 AND 120)
    OR (reaction_type<>'text' AND text_content='')
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
