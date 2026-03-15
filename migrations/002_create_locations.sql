CREATE TABLE locations (
  id          CHAR(36)     NOT NULL DEFAULT (UUID()),
  name        VARCHAR(100) NOT NULL,
  description TEXT,
  latitude    DOUBLE,
  longitude   DOUBLE,
  radius_m    INT          NOT NULL DEFAULT 200,
  is_default  TINYINT(1)   NOT NULL DEFAULT 1,
  created_at  DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_locations_latlng ON locations(latitude, longitude);