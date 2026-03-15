package model

import "time"

// DB レコード
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`           // JSON に出力しない
    DisplayName  string    `json:"display_name"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// リクエスト型
type SignupRequest struct {
    Email       string `json:"email"        binding:"required,email"`
    Password    string `json:"password"     binding:"required,min=8"`
    DisplayName string `json:"display_name" binding:"required"`
}

type LoginRequest struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// レスポンス型
type AuthResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"` // 秒
}

type MeResponse struct {
    ID          string    `json:"id"`
    Email       string    `json:"email"`
    DisplayName string    `json:"display_name"`
    CreatedAt   time.Time `json:"created_at"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}