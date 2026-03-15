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