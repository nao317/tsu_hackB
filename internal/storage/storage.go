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