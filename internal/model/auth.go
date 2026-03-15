package model

import "time"

// User はDBのusersテーブルを表す
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
}

// SignupRequest はサインアップAPIのリクエストボディ
type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

// LoginRequest はログインAPIのリクエストボディ
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse はサインアップ・ログイン・リフレッシュAPIのレスポンス
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshRequest はトークンリフレッシュAPIのリクエストボディ
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest はログアウトAPIのリクエストボディ
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// MeResponse はユーザー情報取得APIのレスポンス
type MeResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}
