CREATE TABLE IF NOT EXISTS refresh_tokens (
  token      VARCHAR(512) NOT NULL,
  user_id    CHAR(36)     NOT NULL,
  expires_at DATETIME     NOT NULL,
  created_at DATETIME     DEFAULT NOW(),
  PRIMARY KEY (token),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
