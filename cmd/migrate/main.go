package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strings"

    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load()

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL が設定されていません")
    }

    // multiStatements=true で1ファイル複数SQL文を許可
    if !strings.Contains(dsn, "multiStatements=true") {
        if strings.Contains(dsn, "?") {
            dsn += "&multiStatements=true"
        } else {
            dsn += "?multiStatements=true"
        }
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