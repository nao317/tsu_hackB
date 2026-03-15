# Backend 実装計画書

## 概要

本ドキュメントは `tsu_hackB`（AACアプリのGoバックエンド）を段階的に実装するための計画書。
設計の詳細は `ARCHITECTURE_README.md` を参照。

### 現状
- `main.go` はスタブ（ルート2つのみ）
- `go.mod` には `gin` のみ登録済み
- ディレクトリ構成・実装なし

### 完成形のディレクトリ構成

```
backend/
├── cmd/
│   ├── server/
│   │   └── main.go          # サーバー起動エントリーポイント
│   └── migrate/
│       └── main.go          # マイグレーション実行CLI
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   └── db.go
│   ├── handler/
│   │   ├── auth.go
│   │   ├── location.go
│   │   ├── card.go
│   │   ├── user_location.go
│   │   └── ai.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── cors.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── location_service.go
│   │   ├── card_service.go
│   │   └── ai_service.go
│   ├── storage/
│   │   ├── storage.go
│   │   └── supabase.go
│   ├── model/
│   │   ├── user.go
│   │   ├── location.go
│   │   ├── card.go
│   │   └── ai.go
│   └── router/
│       └── router.go
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_locations.sql
│   ├── 003_create_cards.sql
│   ├── 004_create_location_cards.sql
│   ├── 005_create_user_locations.sql
│   ├── 006_create_user_location_cards.sql
│   └── 007_seed_initial_data.sql
├── main.go                  # 削除予定（cmd/server/main.go へ移行）
├── Dockerfile               # 後でビルドパス更新が必要
├── docker-compose.yml
├── go.mod
├── .env.example
└── IMPLEMENTATION_PLAN.md
```

---

## 実装ステップ一覧

| Step | 内容 | 依存 |
|------|------|------|
| 1 | 基盤整備（依存パッケージ・config・db・migrations） | なし |
| 2 | モデル層（`internal/model/`） | なし |
| 3 | ストレージ層（`internal/storage/`） | Step 1, 2 |
| 4 | ミドルウェア（`internal/middleware/`） | Step 1, 2 |
| 5 | サービス層（`internal/service/`） | Step 1〜4 |
| 6 | ハンドラ層（`internal/handler/`） | Step 2, 5 |
| 7 | ルーティング・エントリーポイント | Step 4〜6 |

---

## Step 1: 基盤整備

### 1-1. 追加する依存パッケージ

```bash
cd backend

# MySQL ドライバ
go get github.com/go-sql-driver/mysql

# JWT
go get github.com/golang-jwt/jwt/v5

# .env 読み込み
go get github.com/joho/godotenv

# CORS ミドルウェア
go get github.com/gin-contrib/cors

# Supabase Storage（REST APIクライアント）
go get github.com/supabase-community/storage-go
```

> **Gemini API について**
> `google.golang.org/genai` は現時点でアルファ版のため、
> `net/http` で Gemini REST API を直接呼び出す実装を採用する。

### 1-2. `internal/config/config.go`

環境変数をまとめて読み込む構造体。`godotenv` で `.env` ファイルを自動ロード。

```go
package config

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    Port string
    Env  string

    DatabaseURL string

    JWTSecret             string
    JWTAccessExpireMin    int
    JWTRefreshExpireDays  int

    SupabaseURL           string
    SupabaseServiceKey    string
    SupabaseStorageBucket string

    GeminiAPIKey string

    SMTPHost string
    SMTPPort string

    CORSAllowedOrigins string
}

func Load() *Config {
    // .env が存在する場合のみロード（本番では環境変数を直接設定）
    _ = godotenv.Load()

    return &Config{
        Port:                  getEnv("PORT", "8080"),
        Env:                   getEnv("ENV", "development"),
        DatabaseURL:           mustGetEnv("DATABASE_URL"),
        JWTSecret:             mustGetEnv("JWT_SECRET"),
        JWTAccessExpireMin:    getEnvInt("JWT_ACCESS_EXPIRES_MIN", 15),
        JWTRefreshExpireDays:  getEnvInt("JWT_REFRESH_EXPIRES_DAYS", 7),
        SupabaseURL:           mustGetEnv("SUPABASE_URL"),
        SupabaseServiceKey:    mustGetEnv("SUPABASE_SERVICE_KEY"),
        SupabaseStorageBucket: getEnv("SUPABASE_STORAGE_BUCKET", "cards"),
        GeminiAPIKey:          mustGetEnv("GEMINI_API_KEY"),
        SMTPHost:              getEnv("SMTP_HOST", "localhost"),
        SMTPPort:              getEnv("SMTP_PORT", "1025"),
        CORSAllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
    }
}

func getEnv(key, defaultVal string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultVal
}

func mustGetEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatalf("環境変数 %s が設定されていません", key)
    }
    return v
}

func getEnvInt(key string, defaultVal int) int {
    v := os.Getenv(key)
    if v == "" {
        return defaultVal
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        return defaultVal
    }
    return i
}
```

### 1-3. `internal/db/db.go`

```go
package db

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

func New(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("db.Open: %w", err)
    }

    // 接続プール設定
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db.Ping: %w", err)
    }

    return db, nil
}
```

### 1-4. `.env.example` の更新

現在の `.env.example` は PostgreSQL 向けになっているため MySQL 用に修正する。

```env
# サーバー
PORT=8080
ENV=development   # development | production

# MySQL
# Phase 1: Docker MySQL
# Phase 2: Render MySQL 接続文字列に変更
DATABASE_URL=aac:password@tcp(db:3306)/aac_dev?charset=utf8mb4&parseTime=True

# JWT
JWT_SECRET=your-secret-key-min-32chars
JWT_ACCESS_EXPIRES_MIN=15
JWT_REFRESH_EXPIRES_DAYS=7

# Supabase Storage（Phase 1・2 共通）
SUPABASE_URL=https://xxxx.supabase.co
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_STORAGE_BUCKET=cards

# Gemini
GEMINI_API_KEY=AIzaSy...

# SMTP（Phase 1: Mailpit / Phase 2: SendGrid 等）
SMTP_HOST=localhost
SMTP_PORT=1025

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-app.vercel.app
```

### 1-5. `migrations/` SQLファイル

7ファイルを番号順に作成する。

**001_create_users.sql**
```sql
CREATE TABLE users (
  id            CHAR(36)     NOT NULL DEFAULT (UUID()) COMMENT 'ユーザーID',
  email         VARCHAR(255) NOT NULL                  COMMENT 'メールアドレス',
  password_hash VARCHAR(255) NOT NULL                  COMMENT 'bcrypt ハッシュ',
  display_name  VARCHAR(100) NOT NULL                  COMMENT '表示名',
  created_at    DATETIME     NOT NULL DEFAULT NOW(),
  updated_at    DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**002_create_locations.sql**
```sql
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
```

**003_create_cards.sql**
```sql
CREATE TABLE cards (
  id         CHAR(36)     NOT NULL DEFAULT (UUID()),
  label      VARCHAR(100) NOT NULL,
  image_url  TEXT,
  emoji      VARCHAR(10),
  category   VARCHAR(50),
  is_daily   TINYINT(1)   NOT NULL DEFAULT 0,
  created_by CHAR(36),
  created_at DATETIME     NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  FOREIGN KEY fk_cards_user (created_by)
    REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_cards_is_daily    ON cards(is_daily);
CREATE INDEX idx_cards_created_by  ON cards(created_by);
```

**004_create_location_cards.sql**
```sql
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
```

**005_create_user_locations.sql**
```sql
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
```

**006_create_user_location_cards.sql**
```sql
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
```

**007_seed_initial_data.sql**
```sql
-- 共有ロケーションのシード
INSERT INTO locations (id, name, radius_m, is_default) VALUES
  (UUID(), 'コンビニ', 100, 1),
  (UUID(), '病院',     200, 1),
  (UUID(), 'カフェ',   100, 1);

-- 日常カードのシード
INSERT INTO cards (id, label, emoji, is_daily) VALUES
  (UUID(), 'こんにちは', '👋', 1),
  (UUID(), 'ありがとう', '🙏', 1),
  (UUID(), 'すみません', '🙇', 1),
  (UUID(), 'はい',       '✅', 1),
  (UUID(), 'いいえ',     '❌', 1),
  (UUID(), 'おねがい',   '🙏', 1),
  (UUID(), 'わかった',   '👌', 1);
```

### 1-6. `cmd/migrate/main.go`

マイグレーションを番号順に実行するシンプルなCLI。

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"

    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load()

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL が設定されていません")
    }

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("DB接続エラー: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("DB疎通確認エラー: %v", err)
    }

    files, err := filepath.Glob("migrations/*.sql")
    if err != nil || len(files) == 0 {
        log.Fatal("migrations/*.sql が見つかりません")
    }
    sort.Strings(files)

    for _, f := range files {
        fmt.Printf("実行中: %s\n", f)
        content, err := os.ReadFile(f)
        if err != nil {
            log.Fatalf("%s の読み込みエラー: %v", f, err)
        }
        if _, err := db.Exec(string(content)); err != nil {
            log.Fatalf("%s の実行エラー: %v", f, err)
        }
        fmt.Printf("完了: %s\n", f)
    }

    fmt.Println("全マイグレーション完了")
}
```

> **実行方法:**
> ```bash
> docker compose exec backend go run ./cmd/migrate
> ```

### Step 1 完了確認

```bash
# パッケージ追加後
go mod tidy

# ビルドが通るか確認
go build ./...
```

---

## Step 2: モデル層

DBから取得したデータ・リクエスト・レスポンスを表現する構造体群。
ビジネスロジックは含めない。

### `internal/model/user.go`

```go
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
```

### `internal/model/location.go`

```go
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
```

### `internal/model/card.go`

```go
package model

import "time"

// DB レコード
type Card struct {
    ID        string    `json:"id"`
    Label     string    `json:"label"`
    ImageURL  *string   `json:"image_url"`
    Emoji     *string   `json:"emoji"`
    Category  *string   `json:"category"`
    IsDaily   bool      `json:"is_daily"`
    CreatedBy *string   `json:"created_by"`
    CreatedAt time.Time `json:"created_at"`
}

// POST /user/cards のリクエスト（multipart/form-data）
// file フィールドは handler で gin.Context.FormFile() で取得
type CreateCardRequest struct {
    Label    string `form:"label"    binding:"required"`
    Emoji    string `form:"emoji"`
    Category string `form:"category"`
}

// POST /user/locations/:id/cards のリクエスト
type AddCardToLocationRequest struct {
    CardID    string `json:"card_id"    binding:"required"`
    SortOrder int    `json:"sort_order"`
}

// PUT /user/locations/:id/cards/reorder のリクエスト
type ReorderCardsRequest struct {
    Cards []CardOrder `json:"cards" binding:"required"`
}

type CardOrder struct {
    CardID    string `json:"card_id"    binding:"required"`
    SortOrder int    `json:"sort_order" binding:"required"`
}
```

### `internal/model/ai.go`

```go
package model

// POST /ai/recommend のリクエスト
type AIRecommendRequest struct {
    Words        []string `json:"words"         binding:"required,min=1"`
    LocationName string   `json:"location_name"`
}

// POST /ai/recommend のレスポンス
type AIRecommendResponse struct {
    Suggestions []string `json:"suggestions"`
    LatencyMS   int64    `json:"latency_ms"`
}
```

---

## Step 3: ストレージ層

### `internal/storage/storage.go`

```go
package storage

import "context"

// ImageStorage はストレージバックエンドを抽象化するインターフェース。
// 将来的なストレージ先変更はこのインターフェースの実装を差し替えるだけで完結する。
type ImageStorage interface {
    // Upload は画像データをアップロードし、公開URLを返す。
    Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
    // Delete はキーを指定して画像を削除する。
    Delete(ctx context.Context, key string) error
    // GetPublicURL はキーから公開URLを返す（アップロード済み前提）。
    GetPublicURL(key string) string
}
```

### `internal/storage/supabase.go`

```go
package storage

import (
    "bytes"
    "context"
    "fmt"

    storage_go "github.com/supabase-community/storage-go"
)

type SupabaseStorage struct {
    client *storage_go.Client
    bucket string
    baseURL string
}

func NewSupabaseStorage(supabaseURL, serviceKey, bucket string) *SupabaseStorage {
    client := storage_go.NewClient(supabaseURL+"/storage/v1", serviceKey, nil)
    return &SupabaseStorage{
        client:  client,
        bucket:  bucket,
        baseURL: supabaseURL,
    }
}

func (s *SupabaseStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
    _, err := s.client.UploadFile(s.bucket, key, bytes.NewReader(data), storage_go.FileOptions{
        ContentType: &contentType,
    })
    if err != nil {
        return "", fmt.Errorf("supabase upload: %w", err)
    }
    return s.GetPublicURL(key), nil
}

func (s *SupabaseStorage) Delete(ctx context.Context, key string) error {
    _, err := s.client.RemoveFile(s.bucket, []string{key})
    if err != nil {
        return fmt.Errorf("supabase delete: %w", err)
    }
    return nil
}

func (s *SupabaseStorage) GetPublicURL(key string) string {
    return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, key)
}
```

---

## Step 4: ミドルウェア

### `internal/middleware/cors.go`

```go
package middleware

import (
    "strings"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
    origins := strings.Split(allowedOrigins, ",")
    for i := range origins {
        origins[i] = strings.TrimSpace(origins[i])
    }

    return cors.New(cors.Config{
        AllowOrigins:     origins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

### `internal/middleware/auth.go`

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// JWTAuth は Authorization: Bearer <token> を検証し、
// 成功時に c.Set("user_id", userID) をセットするミドルウェア。
func JWTAuth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "認証が必要です",
                "code":  "UNAUTHORIZED",
            })
            return
        }

        tokenStr := strings.TrimPrefix(header, "Bearer ")

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "無効なトークンです",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "トークンの解析に失敗しました",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        userID, ok := claims["sub"].(string)
        if !ok || userID == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "ユーザーIDが取得できません",
                "code":  "INVALID_TOKEN",
            })
            return
        }

        c.Set("user_id", userID)
        c.Next()
    }
}
```

---

## Step 5: サービス層

ビジネスロジックを集約する層。ハンドラからは直接DBやAPIを呼ばず、必ずサービスを経由する。

### `internal/service/auth_service.go`

```go
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
    now := time.Now()
    accessExp := now.Add(time.Duration(s.accessExpireMin) * time.Minute)

    // アクセストークン
    accessClaims := jwt.MapClaims{
        "sub": userID,
        "exp": accessExp.Unix(),
        "iat": now.Unix(),
        "type": "access",
    }
    accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
    if err != nil {
        return nil, fmt.Errorf("sign access token: %w", err)
    }

    // リフレッシュトークン（有効期限は長め）
    refreshExp := now.Add(time.Duration(s.refreshExpireDays) * 24 * time.Hour)
    refreshClaims := jwt.MapClaims{
        "sub":  userID,
        "exp":  refreshExp.Unix(),
        "iat":  now.Unix(),
        "type": "refresh",
    }
    refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
    if err != nil {
        return nil, fmt.Errorf("sign refresh token: %w", err)
    }

    // リフレッシュトークンをDBに保存（ログアウト時に削除して無効化できるようにする）
    _, err = s.db.ExecContext(ctx,
        `INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES (?, ?, ?)
         ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
        refreshToken, userID, refreshExp,
    )
    if err != nil {
        return nil, fmt.Errorf("save refresh token: %w", err)
    }

    return &model.AuthResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "bearer",
        ExpiresIn:    s.accessExpireMin * 60,
    }, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
    // DBにリフレッシュトークンが存在するか確認
    var userID string
    err := s.db.QueryRowContext(ctx,
        "SELECT user_id FROM refresh_tokens WHERE token = ? AND expires_at > NOW()", refreshToken,
    ).Scan(&userID)
    if err == sql.ErrNoRows {
        return nil, ErrInvalidToken
    }
    if err != nil {
        return nil, fmt.Errorf("refresh query: %w", err)
    }

    // 古いトークンを削除してローテーション
    s.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token = ?", refreshToken)

    return s.issueTokens(ctx, userID)
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
```

> **注意:** `refresh_tokens` テーブルが必要。`migrations/` に追加 DDL を作成すること。
> ```sql
> -- migrations/008_create_refresh_tokens.sql
> CREATE TABLE refresh_tokens (
>   token      VARCHAR(512) NOT NULL,
>   user_id    CHAR(36)     NOT NULL,
>   expires_at DATETIME     NOT NULL,
>   PRIMARY KEY (token),
>   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
> ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
> ```

### `internal/service/location_service.go`

```go
package service

import (
    "context"
    "database/sql"
    "fmt"
    "math"
    "sort"

    "github.com/nao317/tsu_hack/backend/internal/model"
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
```

### `internal/service/card_service.go`

```go
package service

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"

    "github.com/nao317/tsu_hack/backend/internal/model"
    "github.com/nao317/tsu_hack/backend/internal/storage"
)

const maxImageSize = 5 * 1024 * 1024 // 5MB

var allowedContentTypes = map[string]bool{
    "image/jpeg": true,
    "image/png":  true,
    "image/webp": true,
}

var ErrInvalidImageType = errors.New("画像はJPEG・PNG・WebPのみ対応しています")
var ErrImageTooLarge    = errors.New("画像サイズは5MB以下にしてください")

type CardService struct {
    db      *sql.DB
    storage storage.ImageStorage
}

func NewCardService(db *sql.DB, storage storage.ImageStorage) *CardService {
    return &CardService{db: db, storage: storage}
}

func (s *CardService) GetDailyCards(ctx context.Context) ([]model.Card, error) {
    rows, err := s.db.QueryContext(ctx,
        "SELECT id, label, image_url, emoji, category, is_daily, created_by, created_at FROM cards WHERE is_daily = 1",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return scanCards(rows)
}

// CreateCard は画像アップロードとDBへのカード保存を1リクエストで完結させる。
func (s *CardService) CreateCard(ctx context.Context, userID string, req *model.CreateCardRequest, file multipart.File, header *multipart.FileHeader) (*model.Card, error) {
    var imageURL *string

    if file != nil {
        // ファイルサイズチェック
        if header.Size > maxImageSize {
            return nil, ErrImageTooLarge
        }

        // Content-Type 検証（先頭512バイトで判定）
        buf := make([]byte, 512)
        n, _ := file.Read(buf)
        ct := http.DetectContentType(buf[:n])
        if !allowedContentTypes[ct] {
            return nil, ErrInvalidImageType
        }

        // ファイル先頭に戻してから全読み込み
        file.Seek(0, io.SeekStart)
        data, err := io.ReadAll(file)
        if err != nil {
            return nil, fmt.Errorf("read file: %w", err)
        }

        // ファイル名はUUIDにリネーム（パストラバーサル防止）
        ext := ".jpg"
        switch ct {
        case "image/png":
            ext = ".png"
        case "image/webp":
            ext = ".webp"
        }

        var key string
        s.db.QueryRowContext(ctx, "SELECT UUID()").Scan(&key)
        key = key + ext

        url, err := s.storage.Upload(ctx, key, data, ct)
        if err != nil {
            return nil, fmt.Errorf("storage upload: %w", err)
        }
        imageURL = &url
    }

    var cardID string
    s.db.QueryRowContext(ctx, "SELECT UUID()").Scan(&cardID)

    _, err := s.db.ExecContext(ctx,
        "INSERT INTO cards (id, label, image_url, emoji, category, created_by) VALUES (?, ?, ?, ?, ?, ?)",
        cardID, req.Label, imageURL, nullStr(req.Emoji), nullStr(req.Category), userID,
    )
    if err != nil {
        // 画像アップロード済みの場合はロールバック
        if imageURL != nil {
            s.storage.Delete(ctx, *imageURL)
        }
        return nil, fmt.Errorf("insert card: %w", err)
    }

    var card model.Card
    s.db.QueryRowContext(ctx,
        "SELECT id, label, image_url, emoji, category, is_daily, created_by, created_at FROM cards WHERE id = ?", cardID,
    ).Scan(&card.ID, &card.Label, &card.ImageURL, &card.Emoji, &card.Category, &card.IsDaily, &card.CreatedBy, &card.CreatedAt)

    return &card, nil
}

func nullStr(s string) interface{} {
    if s == "" {
        return nil
    }
    return s
}
```

### `internal/service/ai_service.go`

```go
package service

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/nao317/tsu_hack/backend/internal/model"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

type AIService struct {
    apiKey     string
    httpClient *http.Client
}

func NewAIService(apiKey string) *AIService {
    return &AIService{
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *AIService) Recommend(ctx context.Context, req *model.AIRecommendRequest) (*model.AIRecommendResponse, error) {
    start := time.Now()

    prompt := buildPrompt(req.Words, req.LocationName)

    // Gemini REST API リクエスト
    body := map[string]interface{}{
        "contents": []map[string]interface{}{
            {
                "parts": []map[string]interface{}{
                    {"text": prompt},
                },
            },
        },
        "generationConfig": map[string]interface{}{
            "temperature":     0.7,
            "maxOutputTokens": 256,
        },
    }

    bodyBytes, _ := json.Marshal(body)
    url := fmt.Sprintf("%s?key=%s", geminiEndpoint, s.apiKey)

    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := s.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("gemini request: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        Candidates []struct {
            Content struct {
                Parts []struct {
                    Text string `json:"text"`
                } `json:"parts"`
            } `json:"content"`
        } `json:"candidates"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("gemini decode: %w", err)
    }

    var suggestions []string
    if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
        text := result.Candidates[0].Content.Parts[0].Text
        // 改行区切りの候補文を分割
        for _, line := range strings.Split(text, "\n") {
            line = strings.TrimSpace(strings.TrimLeft(line, "1234567890.-) "))
            if line != "" {
                suggestions = append(suggestions, line)
            }
        }
    }

    return &model.AIRecommendResponse{
        Suggestions: suggestions,
        LatencyMS:   time.Since(start).Milliseconds(),
    }, nil
}

func buildPrompt(words []string, locationName string) string {
    joined := strings.Join(words, "、")
    location := ""
    if locationName != "" {
        location = fmt.Sprintf("（場所: %s）", locationName)
    }
    return fmt.Sprintf(
        `AAC（拡大代替コミュニケーション）アプリのユーザーが次の単語を選択しました%s。
これらの単語を使って自然で丁寧な日本語文章を2〜3候補生成してください。
助詞（を・に・は・が等）を適切に補完してください。
候補のみを番号付きで出力してください。余計な説明は不要です。

選択単語: %s`, location, joined,
    )
}
```

---

## Step 6: ハンドラ層

ハンドラは薄く保つ。処理フローは「バリデーション → サービス呼び出し → レスポンス」のみ。

### `internal/handler/auth.go`

```go
package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/model"
    "github.com/nao317/tsu_hack/backend/internal/service"
)

type AuthHandler struct {
    svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var req model.SignupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Signup(c.Request.Context(), &req)
    if errors.Is(err, service.ErrEmailAlreadyExists) {
        c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "EMAIL_ALREADY_EXISTS"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Login(c.Request.Context(), &req)
    if errors.Is(err, service.ErrInvalidCredentials) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "INVALID_CREDENTIALS"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var req model.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
    if errors.Is(err, service.ErrInvalidToken) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "INVALID_TOKEN"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
    var req model.LogoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }
    h.svc.Logout(c.Request.Context(), req.RefreshToken)
    c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
    userID := c.GetString("user_id")
    me, err := h.svc.GetMe(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "NOT_FOUND"})
        return
    }
    c.JSON(http.StatusOK, me)
}
```

### `internal/handler/location.go`（抜粋）

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/service"
    "github.com/nao317/tsu_hack/backend/internal/model"
)

type LocationHandler struct {
    svc *service.LocationService
}

func NewLocationHandler(svc *service.LocationService) *LocationHandler {
    return &LocationHandler{svc: svc}
}

func (h *LocationHandler) Nearby(c *gin.Context) {
    var q model.NearbyQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }
    if q.RadiusM == 0 {
        q.RadiusM = 500 // デフォルト500m
    }

    // 認証済みの場合はユーザーロケーションも含める（任意）
    userID, _ := c.Get("user_id")
    uid, _ := userID.(string)

    locs, err := h.svc.GetNearby(c.Request.Context(), q.Lat, q.Lng, q.RadiusM, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, locs)
}

func (h *LocationHandler) List(c *gin.Context) {
    locs, err := h.svc.ListShared(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, locs)
}

func (h *LocationHandler) GetCards(c *gin.Context) {
    locationID := c.Param("id")
    cards, err := h.svc.GetCards(c.Request.Context(), locationID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラー", "code": "INTERNAL_ERROR"})
        return
    }
    c.JSON(http.StatusOK, cards)
}
```

### `internal/handler/card.go`（抜粋）

```go
func (h *CardHandler) CreateCard(c *gin.Context) {
    var req model.CreateCardRequest
    if err := c.ShouldBind(&req); err != nil { // multipart なので ShouldBind
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    // 画像ファイルは任意
    file, header, err := c.Request.FormFile("file")
    if err != nil && err.Error() != "http: no such file" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルの読み込みエラー", "code": "INVALID_FILE"})
        return
    }
    if file != nil {
        defer file.Close()
    }

    userID := c.GetString("user_id")
    card, err := h.svc.CreateCard(c.Request.Context(), userID, &req, file, header)
    // ... エラーハンドリング
    c.JSON(http.StatusCreated, card)
}
```

---

## Step 7: ルーティング & エントリーポイント

### `internal/router/router.go`

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/handler"
    "github.com/nao317/tsu_hack/backend/internal/middleware"
)

type Handlers struct {
    Auth         *handler.AuthHandler
    Location     *handler.LocationHandler
    UserLocation *handler.UserLocationHandler
    Card         *handler.CardHandler
    AI           *handler.AIHandler
}

func Setup(r *gin.Engine, h *Handlers, jwtSecret string, allowedOrigins string) {
    r.Use(middleware.CORS(allowedOrigins))

    v1 := r.Group("/api/v1")

    // 認証不要
    auth := v1.Group("/auth")
    {
        auth.POST("/signup",  h.Auth.Signup)
        auth.POST("/login",   h.Auth.Login)
        auth.POST("/refresh", h.Auth.Refresh)
    }

    // ゲスト可（認証オプション）
    v1.GET("/locations/nearby",        h.Location.Nearby)
    v1.GET("/locations",               h.Location.List)
    v1.GET("/locations/:id/cards",     h.Location.GetCards)
    v1.GET("/cards/daily",             h.Card.Daily)
    v1.POST("/ai/recommend",           h.AI.Recommend)

    // 認証必須
    authed := v1.Group("")
    authed.Use(middleware.JWTAuth(jwtSecret))
    {
        authed.POST("/auth/logout", h.Auth.Logout)
        authed.GET("/auth/me",      h.Auth.Me)

        // ユーザーロケーション
        authed.GET("/user/locations",               h.UserLocation.List)
        authed.POST("/user/locations",              h.UserLocation.Create)
        authed.PUT("/user/locations/:id",           h.UserLocation.Update)
        authed.DELETE("/user/locations/:id",        h.UserLocation.Delete)
        authed.GET("/user/locations/:id/cards",     h.UserLocation.GetCards)

        // カード
        authed.POST("/user/cards",                              h.Card.Create)
        authed.POST("/user/locations/:id/cards",                h.Card.AddToLocation)
        authed.DELETE("/user/locations/:id/cards/:card_id",     h.Card.RemoveFromLocation)
        authed.PUT("/user/locations/:id/cards/reorder",         h.Card.Reorder)
    }
}
```

### `cmd/server/main.go`

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/nao317/tsu_hack/backend/internal/config"
    "github.com/nao317/tsu_hack/backend/internal/db"
    "github.com/nao317/tsu_hack/backend/internal/handler"
    "github.com/nao317/tsu_hack/backend/internal/router"
    "github.com/nao317/tsu_hack/backend/internal/service"
    "github.com/nao317/tsu_hack/backend/internal/storage"
)

func main() {
    cfg := config.Load()

    database, err := db.New(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("DB接続失敗: %v", err)
    }
    defer database.Close()

    // ストレージ
    imageStorage := storage.NewSupabaseStorage(
        cfg.SupabaseURL,
        cfg.SupabaseServiceKey,
        cfg.SupabaseStorageBucket,
    )

    // サービス
    authSvc     := service.NewAuthService(database, cfg.JWTSecret, cfg.JWTAccessExpireMin, cfg.JWTRefreshExpireDays)
    locationSvc := service.NewLocationService(database)
    cardSvc     := service.NewCardService(database, imageStorage)
    aiSvc       := service.NewAIService(cfg.GeminiAPIKey)

    // ハンドラ
    handlers := &router.Handlers{
        Auth:         handler.NewAuthHandler(authSvc),
        Location:     handler.NewLocationHandler(locationSvc),
        UserLocation: handler.NewUserLocationHandler(locationSvc),
        Card:         handler.NewCardHandler(cardSvc),
        AI:           handler.NewAIHandler(aiSvc),
    }

    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    router.Setup(r, handlers, cfg.JWTSecret, cfg.CORSAllowedOrigins)

    log.Printf("サーバー起動: :%s", cfg.Port)
    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatalf("サーバー起動失敗: %v", err)
    }
}
```

### Dockerfile の更新

`cmd/server/main.go` にエントリーポイントを移行したため、Dockerfile のビルドコマンドを変更する。

```dockerfile
# 変更前
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 変更後
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server
```

---

## エラーレスポンス形式

全エンドポイントで統一されたエラー形式を使用する。

```json
{
  "error": "エラーの説明（日本語）",
  "code":  "MACHINE_READABLE_CODE"
}
```

| code | HTTPステータス | 意味 |
|------|--------------|------|
| `VALIDATION_ERROR`      | 400 | リクエストのバリデーションエラー |
| `UNAUTHORIZED`          | 401 | 認証が必要 |
| `INVALID_TOKEN`         | 401 | トークンが無効または期限切れ |
| `INVALID_CREDENTIALS`   | 401 | メールアドレスまたはパスワードが不正 |
| `FORBIDDEN`             | 403 | アクセス権限なし |
| `NOT_FOUND`             | 404 | リソースが存在しない |
| `EMAIL_ALREADY_EXISTS`  | 409 | メールアドレス重複 |
| `INVALID_FILE`          | 400 | 画像ファイルの形式・サイズ不正 |
| `INTERNAL_ERROR`        | 500 | サーバー内部エラー |

---

## 進捗チェックリスト

- [ ] **Step 1**: 基盤整備
  - [ ] パッケージ追加 & `go mod tidy`
  - [ ] `internal/config/config.go`
  - [ ] `internal/db/db.go`
  - [ ] `.env.example` 更新（MySQL用）
  - [ ] `migrations/` 7ファイル（+ 008_refresh_tokens.sql）
  - [ ] `cmd/migrate/main.go`
  - [ ] `go build ./...` が通ること
- [ ] **Step 2**: モデル層（`internal/model/`）
- [ ] **Step 3**: ストレージ層（`internal/storage/`）
- [ ] **Step 4**: ミドルウェア（`internal/middleware/`）
- [ ] **Step 5**: サービス層（`internal/service/`）
- [ ] **Step 6**: ハンドラ層（`internal/handler/`）
- [ ] **Step 7**: ルーティング & エントリーポイント（`cmd/server/main.go`）
  - [ ] `Dockerfile` のビルドパス更新
  - [ ] `docker compose up` で起動確認
  - [ ] `curl localhost:8080/health` が `{"status":"ok"}` を返すこと
