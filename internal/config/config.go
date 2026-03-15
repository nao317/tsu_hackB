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