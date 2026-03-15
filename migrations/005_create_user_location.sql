CREATE TABLE user_locations (
  id         CHAR(36)     NOT NULL DEFAULT (UUID()),
  user_id    CHAR(36)     NOT NULL,
  name       VARCHAR(100) NOT NULL,
  latitude   DOUBLE       NOT NULL,
  longitude  DOUBLE       NOT NULL,
  radius_m   INT          NOT NULL DEFAULT 100,
  created_at DATETIME     NOT NULL DEFAULT NOW(),
  updated_at DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  FOREIGN KEY fk_ul_user (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_user_locations_user   ON user_locations(user_id);
CREATE INDEX idx_user_locations_latlng ON user_locations(latitude, longitude);