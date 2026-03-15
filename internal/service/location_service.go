package service

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "math"
    "sort"

    "github.com/nao317/tsu_hack/backend/internal/model"
)

var (
    ErrNotFound  = errors.New("見つかりません")
    ErrForbidden = errors.New("権限がありません")
)

type LocationService struct {
    db *sql.DB
}

func NewLocationService(db *sql.DB) *LocationService {
    return &LocationService{db: db}
}

// Haversine は2点間の距離をメートルで返す。
func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
    const earthR = 6371000.0 // 地球半径（メートル）
    φ1 := lat1 * math.Pi / 180
    φ2 := lat2 * math.Pi / 180
    Δφ := (lat2 - lat1) * math.Pi / 180
    Δλ := (lng2 - lng1) * math.Pi / 180

    a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
        math.Cos(φ1)*math.Cos(φ2)*
            math.Sin(Δλ/2)*math.Sin(Δλ/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return earthR * c
}

// GetNearby は現在地から radius_m 以内のロケーション（共有＋ユーザー）を距離昇順で返す。
func (s *LocationService) GetNearby(ctx context.Context, lat, lng float64, radiusM int, userID string) ([]model.NearbyLocation, error) {
    const delta = 0.01 // 約1.1km の概算フィルタ

    // 共有ロケーションを概算フィルタで絞り込み
    rows, err := s.db.QueryContext(ctx, `
        SELECT l.id, l.name, l.latitude, l.longitude, l.radius_m,
               COUNT(lc.card_id) AS cards_count
        FROM locations l
        LEFT JOIN location_cards lc ON lc.location_id = l.id
        WHERE l.latitude  BETWEEN ? AND ?
          AND l.longitude BETWEEN ? AND ?
        GROUP BY l.id`,
        lat-delta, lat+delta, lng-delta, lng+delta,
    )
    if err != nil {
        return nil, fmt.Errorf("nearby shared query: %w", err)
    }
    defer rows.Close()

    var results []model.NearbyLocation
    for rows.Next() {
        var loc struct {
            id, name          string
            lat, lng          float64
            locRadiusM        int
            cardsCount        int
        }
        if err := rows.Scan(&loc.id, &loc.name, &loc.lat, &loc.lng, &loc.locRadiusM, &loc.cardsCount); err != nil {
            continue
        }
        dist := Haversine(lat, lng, loc.lat, loc.lng)
        if dist > float64(radiusM) {
            continue
        }
        results = append(results, model.NearbyLocation{
            ID: loc.id, Name: loc.name, Type: "shared",
            DistanceM: dist, CardsCount: loc.cardsCount,
        })
    }

    // ログインユーザーのロケーションも取得
    if userID != "" {
        userRows, err := s.db.QueryContext(ctx, `
            SELECT ul.id, ul.name, ul.latitude, ul.longitude, ul.radius_m,
                   COUNT(ulc.card_id) AS cards_count
            FROM user_locations ul
            LEFT JOIN user_location_cards ulc ON ulc.user_location_id = ul.id
            WHERE ul.user_id = ?
              AND ul.latitude  BETWEEN ? AND ?
              AND ul.longitude BETWEEN ? AND ?
            GROUP BY ul.id`,
            userID, lat-delta, lat+delta, lng-delta, lng+delta,
        )
        if err == nil {
            defer userRows.Close()
            for userRows.Next() {
                var loc struct {
                    id, name          string
                    lat, lng          float64
                    locRadiusM        int
                    cardsCount        int
                }
                if err := userRows.Scan(&loc.id, &loc.name, &loc.lat, &loc.lng, &loc.locRadiusM, &loc.cardsCount); err != nil {
                    continue
                }
                dist := Haversine(lat, lng, loc.lat, loc.lng)
                if dist > float64(loc.locRadiusM) {
                    continue
                }
                results = append(results, model.NearbyLocation{
                    ID: loc.id, Name: loc.name, Type: "user",
                    DistanceM: dist, CardsCount: loc.cardsCount,
                })
            }
        }
    }

    // 距離昇順でソート
    sort.Slice(results, func(i, j int) bool {
        return results[i].DistanceM < results[j].DistanceM
    })

    return results, nil
}

func (s *LocationService) ListShared(ctx context.Context) ([]model.Location, error) {
    rows, err := s.db.QueryContext(ctx,
        "SELECT id, name, description, latitude, longitude, radius_m, is_default, created_at FROM locations",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var locs []model.Location
    for rows.Next() {
        var l model.Location
        if err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.Latitude, &l.Longitude, &l.RadiusM, &l.IsDefault, &l.CreatedAt); err != nil {
            return nil, err
        }
        locs = append(locs, l)
    }
    return locs, nil
}

func (s *LocationService) GetCards(ctx context.Context, locationID string) ([]model.Card, error) {
    rows, err := s.db.QueryContext(ctx, `
        SELECT c.id, c.label, c.image_url, c.emoji, c.category, c.is_daily, c.created_by, c.created_at
        FROM cards c
        JOIN location_cards lc ON lc.card_id = c.id
        WHERE lc.location_id = ?
        ORDER BY lc.sort_order`, locationID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return scanCards(rows)
}

func (s *LocationService) ListUserLocations(ctx context.Context, userID string) ([]model.UserLocation, error) {
    rows, err := s.db.QueryContext(ctx,
        "SELECT id, user_id, name, latitude, longitude, radius_m, created_at, updated_at FROM user_locations WHERE user_id = ?",
        userID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var locs []model.UserLocation
    for rows.Next() {
        var l model.UserLocation
        if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Latitude, &l.Longitude, &l.RadiusM, &l.CreatedAt, &l.UpdatedAt); err != nil {
            return nil, err
        }
        locs = append(locs, l)
    }
    return locs, nil
}

func (s *LocationService) CreateUserLocation(ctx context.Context, userID string, req *model.CreateUserLocationRequest) (*model.UserLocation, error) {
    var id string
    s.db.QueryRowContext(ctx, "SELECT UUID()").Scan(&id)

    radiusM := req.RadiusM
    if radiusM == 0 {
        radiusM = 200
    }

    _, err := s.db.ExecContext(ctx,
        "INSERT INTO user_locations (id, user_id, name, latitude, longitude, radius_m) VALUES (?, ?, ?, ?, ?, ?)",
        id, userID, req.Name, req.Latitude, req.Longitude, radiusM,
    )
    if err != nil {
        return nil, fmt.Errorf("create user location: %w", err)
    }

    var loc model.UserLocation
    s.db.QueryRowContext(ctx,
        "SELECT id, user_id, name, latitude, longitude, radius_m, created_at, updated_at FROM user_locations WHERE id = ?", id,
    ).Scan(&loc.ID, &loc.UserID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.RadiusM, &loc.CreatedAt, &loc.UpdatedAt)
    return &loc, nil
}

func (s *LocationService) UpdateUserLocation(ctx context.Context, id, userID string, req *model.UpdateUserLocationRequest) (*model.UserLocation, error) {
    var ownerID string
    err := s.db.QueryRowContext(ctx, "SELECT user_id FROM user_locations WHERE id = ?", id).Scan(&ownerID)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    if ownerID != userID {
        return nil, ErrForbidden
    }

    _, err = s.db.ExecContext(ctx,
        `UPDATE user_locations SET
            name      = IF(? != '', ?, name),
            latitude  = IF(? != 0, ?, latitude),
            longitude = IF(? != 0, ?, longitude),
            radius_m  = IF(? != 0, ?, radius_m),
            updated_at = NOW()
         WHERE id = ?`,
        req.Name, req.Name,
        req.Latitude, req.Latitude,
        req.Longitude, req.Longitude,
        req.RadiusM, req.RadiusM,
        id,
    )
    if err != nil {
        return nil, fmt.Errorf("update user location: %w", err)
    }

    var loc model.UserLocation
    s.db.QueryRowContext(ctx,
        "SELECT id, user_id, name, latitude, longitude, radius_m, created_at, updated_at FROM user_locations WHERE id = ?", id,
    ).Scan(&loc.ID, &loc.UserID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.RadiusM, &loc.CreatedAt, &loc.UpdatedAt)
    return &loc, nil
}

func (s *LocationService) DeleteUserLocation(ctx context.Context, id, userID string) error {
    var ownerID string
    err := s.db.QueryRowContext(ctx, "SELECT user_id FROM user_locations WHERE id = ?", id).Scan(&ownerID)
    if err == sql.ErrNoRows {
        return ErrNotFound
    }
    if err != nil {
        return err
    }
    if ownerID != userID {
        return ErrForbidden
    }

    _, err = s.db.ExecContext(ctx, "DELETE FROM user_locations WHERE id = ?", id)
    return err
}

func (s *LocationService) GetUserLocationCards(ctx context.Context, locationID, userID string) ([]model.Card, error) {
    var ownerID string
    err := s.db.QueryRowContext(ctx, "SELECT user_id FROM user_locations WHERE id = ?", locationID).Scan(&ownerID)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    if ownerID != userID {
        return nil, ErrForbidden
    }

    rows, err := s.db.QueryContext(ctx, `
        SELECT c.id, c.label, c.image_url, c.emoji, c.category, c.is_daily, c.created_by, c.created_at
        FROM cards c
        JOIN user_location_cards ulc ON ulc.card_id = c.id
        WHERE ulc.user_location_id = ?
        ORDER BY ulc.sort_order`, locationID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return scanCards(rows)
}

func scanCards(rows *sql.Rows) ([]model.Card, error) {
    var cards []model.Card
    for rows.Next() {
        var c model.Card
        if err := rows.Scan(&c.ID, &c.Label, &c.ImageURL, &c.Emoji, &c.Category, &c.IsDaily, &c.CreatedBy, &c.CreatedAt); err != nil {
            return nil, err
        }
        cards = append(cards, c)
    }
    return cards, nil
}