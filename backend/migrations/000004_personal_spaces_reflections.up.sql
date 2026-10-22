ALTER TABLE moments
  MODIFY COLUMN couple_id BIGINT UNSIGNED NULL,
  ADD COLUMN space_type ENUM('shared','personal') NULL AFTER author_user_id,
  ADD COLUMN visibility ENUM('couple','private','partner') NULL AFTER space_type,
  ADD COLUMN kind ENUM('standard','daily_question') NULL AFTER visibility,
  ADD COLUMN source_moment_id BIGINT UNSIGNED NULL AFTER kind,
  ADD COLUMN daily_question_id BIGINT UNSIGNED NULL AFTER source_moment_id;

ALTER TABLE notification_preferences
  ADD COLUMN ai_personalization_enabled BOOLEAN NOT NULL DEFAULT TRUE AFTER nudge_enabled;

UPDATE moments
SET space_type='shared', visibility='couple', kind='standard'
WHERE space_type IS NULL OR visibility IS NULL OR kind IS NULL;

ALTER TABLE moments
  MODIFY COLUMN space_type ENUM('shared','personal') NOT NULL,
  MODIFY COLUMN visibility ENUM('couple','private','partner') NOT NULL,
  MODIFY COLUMN kind ENUM('standard','daily_question') NOT NULL DEFAULT 'standard',
  ADD CONSTRAINT chk_moment_space_visibility CHECK (
    (space_type='shared' AND visibility='couple' AND couple_id IS NOT NULL)
    OR (space_type='personal' AND visibility IN ('private','partner'))
  ),
  ADD CONSTRAINT chk_moment_kind_links CHECK (
    (kind='standard' AND daily_question_id IS NULL)
    OR (kind='daily_question' AND space_type='shared' AND daily_question_id IS NOT NULL)
  ),
  ADD CONSTRAINT fk_moment_source FOREIGN KEY (source_moment_id) REFERENCES moments(id) ON DELETE SET NULL,
  ADD KEY idx_moments_personal_timeline (author_user_id, space_type, visibility, occurred_at DESC, id DESC),
  ADD KEY idx_moments_source (source_moment_id);

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

ALTER TABLE moments
  ADD CONSTRAINT fk_moment_daily_question FOREIGN KEY (daily_question_id) REFERENCES daily_questions(id),
  ADD UNIQUE KEY uq_moment_daily_question (daily_question_id);

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
