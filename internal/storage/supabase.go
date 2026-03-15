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