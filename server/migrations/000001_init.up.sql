CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  wechat_openid VARCHAR(128) NOT NULL,
  display_name VARCHAR(120) NOT NULL DEFAULT '',
  avatar_ref VARCHAR(1024) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id), UNIQUE KEY uq_users_openid (wechat_openid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE couples (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_a_id BIGINT UNSIGNED NOT NULL,
  user_b_id BIGINT UNSIGNED NOT NULL,
  status ENUM('active','ended') NOT NULL DEFAULT 'active',
  bound_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  ended_at DATETIME(6) NULL,
  active_user_a BIGINT UNSIGNED GENERATED ALWAYS AS (CASE WHEN status='active' THEN user_a_id ELSE NULL END) STORED,
  active_user_b BIGINT UNSIGNED GENERATED ALWAYS AS (CASE WHEN status='active' THEN user_b_id ELSE NULL END) STORED,
  PRIMARY KEY (id),
  CONSTRAINT chk_couple_order CHECK (user_a_id < user_b_id),
  CONSTRAINT fk_couple_a FOREIGN KEY (user_a_id) REFERENCES users(id),
  CONSTRAINT fk_couple_b FOREIGN KEY (user_b_id) REFERENCES users(id),
  UNIQUE KEY uq_active_user_a (active_user_a), UNIQUE KEY uq_active_user_b (active_user_b),
  KEY idx_couple_a_status (user_a_id, status), KEY idx_couple_b_status (user_b_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE invite_codes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, code CHAR(6) NOT NULL,
  creator_user_id BIGINT UNSIGNED NOT NULL, expires_at DATETIME(6) NOT NULL,
  consumed_at DATETIME(6) NULL, consumer_user_id BIGINT UNSIGNED NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id), UNIQUE KEY uq_invite_code (code),
  CONSTRAINT fk_invite_creator FOREIGN KEY (creator_user_id) REFERENCES users(id),
  CONSTRAINT fk_invite_consumer FOREIGN KEY (consumer_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, couple_id BIGINT UNSIGNED NOT NULL,
  author_user_id BIGINT UNSIGNED NOT NULL, body VARCHAR(2000) NULL,
  status_code VARCHAR(32) NULL, place_name VARCHAR(320) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id), KEY idx_moments_timeline (couple_id, created_at DESC, id DESC),
  CONSTRAINT fk_moment_couple FOREIGN KEY (couple_id) REFERENCES couples(id),
  CONSTRAINT fk_moment_author FOREIGN KEY (author_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE moment_photos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, moment_id BIGINT UNSIGNED NOT NULL,
  object_key VARCHAR(1024) NOT NULL, sort_order TINYINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), PRIMARY KEY (id),
  UNIQUE KEY uq_moment_photo_order (moment_id, sort_order),
  CONSTRAINT chk_photo_order CHECK (sort_order BETWEEN 1 AND 9),
  CONSTRAINT fk_photo_moment FOREIGN KEY (moment_id) REFERENCES moments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, token_hash BINARY(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL, expires_at DATETIME(6) NOT NULL,
  revoked_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id), UNIQUE KEY uq_refresh_hash (token_hash), KEY idx_refresh_user (user_id),
  CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE upload_grants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, user_id BIGINT UNSIGNED NOT NULL,
  object_key VARCHAR(1024) NOT NULL, content_type VARCHAR(64) NOT NULL,
  expires_at DATETIME(6) NOT NULL, consumed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), PRIMARY KEY (id),
  UNIQUE KEY uq_upload_object (object_key), KEY idx_upload_user (user_id, expires_at),
  CONSTRAINT fk_upload_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
