CREATE TABLE user_location_cards (
  id               CHAR(36) NOT NULL DEFAULT (UUID()),
  user_location_id CHAR(36) NOT NULL,
  card_id          CHAR(36) NOT NULL,
  sort_order       INT      NOT NULL DEFAULT 0,
  created_at       DATETIME NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE KEY uq_ulc (user_location_id, card_id),
  FOREIGN KEY fk_ulc_location (user_location_id) REFERENCES user_locations(id) ON DELETE CASCADE,
  FOREIGN KEY fk_ulc_card     (card_id)           REFERENCES cards(id)          ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_ulc_ul ON user_location_cards(user_location_id, sort_order);