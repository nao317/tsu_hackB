CREATE TABLE users (
  id            CHAR(36)     NOT NULL DEFAULT (UUID()) COMMENT 'ユーザーID',
  email         VARCHAR(255) NOT NULL                  COMMENT 'メールアドレス',
  password_hash VARCHAR(255) NOT NULL                  COMMENT 'bcrypt ハッシュ',
  display_name  VARCHAR(100) NOT NULL                  COMMENT '表示名',
  created_at    DATETIME     NOT NULL DEFAULT NOW(),
  updated_at    DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;