# --- Stage 1: Build stage ---
FROM golang:1.26.1-alpine AS builder

# ビルドに必要なツール（gitなど）をインストール
RUN apk add --no-cache git

WORKDIR /app

# 依存関係をコピー、キャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .
# ビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

# 実行用のイメージ指定
FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/server /server

# root権限での実行を拒否
USER nonroot:nonroot

ENTRYPOINT ["/server"]
