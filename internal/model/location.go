package model

import "time"

// 共有ロケーション（DBレコード）
type Location struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description *string   `json:"description"`
    Latitude    *float64  `json:"latitude"`
    Longitude   *float64  `json:"longitude"`
    RadiusM     int       `json:"radius_m"`
    IsDefault   bool      `json:"is_default"`
    CreatedAt   time.Time `json:"created_at"`
}

// GET /locations/nearby のレスポンス要素
type NearbyLocation struct {
    ID         string  `json:"id"`
    Name       string  `json:"name"`
    Type       string  `json:"type"`       // "shared" | "user"
    DistanceM  float64 `json:"distance_m"`
    CardsCount int     `json:"cards_count"`
}

// ユーザー独自ロケーション（DBレコード）
type UserLocation struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Name      string    `json:"name"`
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    RadiusM   int       `json:"radius_m"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// POST/PUT /user/locations のリクエスト
type CreateUserLocationRequest struct {
    Name      string  `json:"name"      binding:"required"`
    Latitude  float64 `json:"latitude"  binding:"required"`
    Longitude float64 `json:"longitude" binding:"required"`
    RadiusM   int     `json:"radius_m"`
}

type UpdateUserLocationRequest struct {
    Name      string  `json:"name"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    RadiusM   int     `json:"radius_m"`
}

// GET /locations/nearby のクエリパラメータ
type NearbyQuery struct {
    Lat      float64 `form:"lat"      binding:"required"`
    Lng      float64 `form:"lng"      binding:"required"`
    RadiusM  int     `form:"radius_m"`
}