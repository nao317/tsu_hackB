CREATE TABLE location_cards (
  id          CHAR(36) NOT NULL DEFAULT (UUID()),
  location_id CHAR(36) NOT NULL,
  card_id     CHAR(36) NOT NULL,
  sort_order  INT      NOT NULL DEFAULT 0,
  created_at  DATETIME NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE KEY uq_loc_card (location_id, card_id),
  FOREIGN KEY fk_lc_location (location_id) REFERENCES locations(id) ON DELETE CASCADE,
  FOREIGN KEY fk_lc_card    (card_id)     REFERENCES cards(id)     ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_location_cards_loc ON location_cards(location_id, sort_order);