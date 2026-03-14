# --- Stage 1: Build stage ---
FROM golang:1.26.1-alpine AS builder

# ビルドに必要なツール（gitなど）をインストール
RUN apk add --no-cache git

WORKDIR /app

# 依存関係を先にコピーしてキャッシュを効かせる
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- Stage 2: Final stage ---
FROM alpine:latest

# 実行に必要な最小限のライブラリ（CA証明書など）をインストール
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドステージから実行バイナリのみをコピー
COPY --from=builder /app/main .
# もし config ファイルや静的ファイルがあればコピー
# COPY --from=builder /app/config ./config

# コンテナ起動時に実行
CMD ["./main"]
