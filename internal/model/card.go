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