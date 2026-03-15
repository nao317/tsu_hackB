package service

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/nao317/tsu_hack/backend/internal/model"
    "golang.org/x/crypto/bcrypt"
)

var (
    ErrEmailAlreadyExists = errors.New("このメールアドレスは既に使用されています")
    ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")
    ErrInvalidToken       = errors.New("無効なトークンです")
)

type AuthService struct {
    db                   *sql.DB
    jwtSecret            []byte
    accessExpireMin      int
    refreshExpireDays    int
}

type tokenClaims struct {
    TokenType string `json:"type"`
    jwt.RegisteredClaims
}

type signedTokens struct {
    accessToken  string
    refreshToken string
    refreshExp   time.Time
}

func NewAuthService(db *sql.DB, jwtSecret string, accessExpireMin, refreshExpireDays int) *AuthService {
    return &AuthService{
        db:                db,
        jwtSecret:         []byte(jwtSecret),
        accessExpireMin:   accessExpireMin,
        refreshExpireDays: refreshExpireDays,
    }
}

func (s *AuthService) Signup(ctx context.Context, req *model.SignupRequest) (*model.AuthResponse, error) {
    // メールアドレスの重複確認
    var exists bool
    err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("signup exists check: %w", err)
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }

    // パスワードハッシュ化（コスト12）
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    if err != nil {
        return nil, fmt.Errorf("bcrypt: %w", err)
    }

    // ユーザー登録
    var userID string
    err = s.db.QueryRowContext(ctx, "SELECT UUID()").Scan(&userID)
    if err != nil {
        return nil, err
    }

    _, err = s.db.ExecContext(ctx,
        "INSERT INTO users (id, email, password_hash, display_name) VALUES (?, ?, ?, ?)",
        userID, req.Email, string(hash), req.DisplayName,
    )
    if err != nil {
        return nil, fmt.Errorf("signup insert: %w", err)
    }

    return s.issueTokens(ctx, userID)
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
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

    return s.issueTokens(ctx, user.ID)
}

// issueTokens はアクセストークンとリフレッシュトークンを発行する。
// リフレッシュトークンはDBに保存する。
func (s *AuthService) issueTokens(ctx context.Context, userID string) (*model.AuthResponse, error) {
    tokens, err := s.generateSignedTokens(userID)
    if err != nil {
        return nil, err
    }

    // リフレッシュトークンをDBに保存（ログアウト時に削除して無効化できるようにする）
    _, err = s.db.ExecContext(ctx,
        `INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)
         ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
        tokens.refreshToken, userID, tokens.refreshExp,
    )
    if err != nil {
        return nil, fmt.Errorf("save refresh token: %w", err)
    }

    return &model.AuthResponse{
        AccessToken:  tokens.accessToken,
        RefreshToken: tokens.refreshToken,
        TokenType:    "bearer",
        ExpiresIn:    s.accessExpireMin * 60,
    }, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin refresh tx: %w", err)
    }
    defer tx.Rollback()

    // DBにリフレッシュトークンが存在するか確認し、同時更新を防ぐためロックする
    var userID string
    err = tx.QueryRowContext(ctx,
        "SELECT user_id FROM refresh_tokens WHERE token = ? AND expires_at > NOW() FOR UPDATE", refreshToken,
    ).Scan(&userID)
    if err == sql.ErrNoRows {
        return nil, ErrInvalidToken
    }
    if err != nil {
        return nil, fmt.Errorf("refresh query: %w", err)
    }

    // 古いトークンを削除してローテーション
    if _, err := tx.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token = ?", refreshToken); err != nil {
        return nil, fmt.Errorf("delete old refresh token: %w", err)
    }

    tokens, err := s.generateSignedTokens(userID)
    if err != nil {
        return nil, err
    }

    _, err = tx.ExecContext(ctx,
        `INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)
         ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
        tokens.refreshToken, userID, tokens.refreshExp,
    )
    if err != nil {
        return nil, fmt.Errorf("save refresh token: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit refresh tx: %w", err)
    }

    return &model.AuthResponse{
        AccessToken:  tokens.accessToken,
        RefreshToken: tokens.refreshToken,
        TokenType:    "bearer",
        ExpiresIn:    s.accessExpireMin * 60,
    }, nil
}

func (s *AuthService) generateSignedTokens(userID string) (*signedTokens, error) {
    now := time.Now()
    accessExp := now.Add(time.Duration(s.accessExpireMin) * time.Minute)

    accessClaims := tokenClaims{
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   userID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(accessExp),
        },
    }
    accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
    if err != nil {
        return nil, fmt.Errorf("sign access token: %w", err)
    }

    refreshExp := now.Add(time.Duration(s.refreshExpireDays) * 24 * time.Hour)
    refreshClaims := tokenClaims{
        TokenType: "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   userID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(refreshExp),
        },
    }
    refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
    if err != nil {
        return nil, fmt.Errorf("sign refresh token: %w", err)
    }

    return &signedTokens{
        accessToken:  accessToken,
        refreshToken: refreshToken,
        refreshExp:   refreshExp,
    }, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
    _, err := s.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token = ?", refreshToken)
    return err
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
