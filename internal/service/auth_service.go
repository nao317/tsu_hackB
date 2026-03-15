package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nao317/tsu_hack/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("このメールアドレスは既に使用されています")
	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Signup(ctx context.Context, req *model.SignupRequest) (*model.MeResponse, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("signup exists check: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}

	var userID string
	if err := s.db.QueryRowContext(ctx, "SELECT UUID()").Scan(&userID); err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, display_name) VALUES (?, ?, ?, ?)",
		userID, req.Email, string(hash), req.DisplayName,
	)
	if err != nil {
		return nil, fmt.Errorf("signup insert: %w", err)
	}

	return s.GetMe(ctx, userID)
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.MeResponse, error) {
	var user model.User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, password_hash FROM users WHERE email = ?", req.Email,
	).Scan(&user.ID, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("login query: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.GetMe(ctx, user.ID)
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*model.MeResponse, error) {
	var me model.MeResponse
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, display_name, created_at FROM users WHERE id = ?", userID,
	).Scan(&me.ID, &me.Email, &me.DisplayName, &me.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ユーザーが見つかりません")
	}
	return &me, err
}
