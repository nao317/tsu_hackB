CREATE TABLE cards (
  id         CHAR(36)     NOT NULL DEFAULT (UUID()),
  label      VARCHAR(100) NOT NULL,
  image_url  TEXT,
  emoji      VARCHAR(10),
  category   VARCHAR(50),
  is_daily   TINYINT(1)   NOT NULL DEFAULT 0,
  created_by CHAR(36),
  created_at DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  FOREIGN KEY fk_cards_user (created_by)
    REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_cards_is_daily    ON cards(is_daily);
CREATE INDEX idx_cards_created_by  ON cards(created_by);