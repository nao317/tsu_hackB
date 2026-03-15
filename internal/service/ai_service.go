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