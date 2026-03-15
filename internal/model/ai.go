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