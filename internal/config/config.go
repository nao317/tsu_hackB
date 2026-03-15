package config

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    Port string
    Env  string

    DatabaseURL string

    SupabaseURL           string
    SupabaseServiceKey    string
    SupabaseStorageBucket string

    GeminiAPIKey string

    JWTSecret            string
    JWTAccessExpireMin   int
    JWTRefreshExpireDays int

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
        DatabaseURL:           resolveDatabaseURL(),
        SupabaseURL:           mustGetEnv("SUPABASE_URL"),
        SupabaseServiceKey:    mustGetEnv("SUPABASE_SERVICE_KEY"),
        SupabaseStorageBucket: getEnv("SUPABASE_STORAGE_BUCKET", "cards"),
        GeminiAPIKey:          mustGetEnv("GEMINI_API_KEY"),
        JWTSecret:             mustGetEnv("JWT_SECRET"),
        JWTAccessExpireMin:    getEnvInt("JWT_ACCESS_EXPIRE_MIN", 60),
        JWTRefreshExpireDays:  getEnvInt("JWT_REFRESH_EXPIRE_DAYS", 30),
        SMTPHost:              getEnv("SMTP_HOST", "localhost"),
        SMTPPort:              getEnv("SMTP_PORT", "1025"),
        CORSAllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
    }
}

func resolveDatabaseURL() string {
    if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
        return dsn
    }

    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dbHost := getEnv("DB_HOST", "localhost")
    dbName := os.Getenv("DB_NAME")
    dbPort := getEnv("DB_PORT", "3306")

    if dbUser == "" || dbPass == "" || dbName == "" {
        log.Fatal("環境変数 DATABASE_URL または DB_USER/DB_PASS/DB_NAME が設定されていません")
    }

    return fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
        dbUser,
        dbPass,
        dbHost,
        dbPort,
        dbName,
    )
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
    raw := os.Getenv(key)
    if raw == "" {
        return defaultVal
    }

    v, err := strconv.Atoi(raw)
    if err != nil {
        log.Printf("環境変数 %s が不正です。デフォルト値 %d を使用します", key, defaultVal)
        return defaultVal
    }
    return v
}

