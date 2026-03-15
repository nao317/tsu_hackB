# バックエンド

Go / Gin による REST API サーバー。

---

## 必要なもの

- Docker / Docker Compose v5+
- Go 1.26.1（ローカルでマイグレーション・ビルドを行う場合）

---

## 環境変数

`docker-compose.yml` の `app` サービスに以下を設定してください。

| 変数名 | 説明 | デフォルト（docker-compose） |
|---|---|---|
| `DATABASE_URL` | MySQL DSN | `user:password@tcp(mysql:3306)/mydb?parseTime=true` |
| `SUPABASE_URL` | Supabase プロジェクト URL | `$SUPABASE_URL` |
| `SUPABASE_SERVICE_KEY` | Supabase サービスロールキー | `$SUPABASE_SERVICE_KEY` |
| `GEMINI_API_KEY` | Gemini API キー | `$GEMINI_API_KEY` |
| `CORS_ALLOWED_ORIGINS` | 許可するオリジン（カンマ区切り） | `http://localhost:3000` |
| `PORT` | サーバーポート | `8080` |
| `ENV` | 実行環境（`development` / `production`） | `development` |

Supabase・Gemini のキーはシェル環境変数にセットするか、`.env` ファイルに書いておくと `docker-compose.yml` の `${変数名}` で自動的に読み込まれます。

```bash
# .env（リポジトリにコミットしないこと）
SUPABASE_URL=https://xxxx.supabase.co
SUPABASE_SERVICE_KEY=eyJ...
GEMINI_API_KEY=AIza...
```

---

## 起動手順

### 1. コンテナをビルドして起動

```bash
sudo docker-compose up -d --build
```

初回はイメージのビルドがあるため数分かかります。

```bash
# 起動確認
sudo docker-compose ps

# アプリログの確認
sudo docker-compose logs -f app
```

### 2. マイグレーション実行

**テーブル作成・初期データ投入はコンテナ外（ローカル）から行います。**

```bash
DATABASE_URL="user:password@tcp(localhost:3306)/mydb?parseTime=true" \
  go run ./cmd/migrate/
```

成功すると以下のように表示されます：

```
実行中: migrations/001_create_users.sql
完了: migrations/001_create_users.sql
...
全マイグレーション完了
```

> **注意:** テーブルがすでに存在する場合はエラーになります。
> DB をリセットしてから再実行してください。
>
> ```bash
> # DB リセット（データが全て消えます）
> sudo docker-compose exec mysql mysql -uroot -proot_password \
>   -e "DROP DATABASE mydb; CREATE DATABASE mydb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
> ```

### 3. 動作確認

```bash
curl http://localhost:8080/health
# => {"status":"ok"}
```

---

## エンドポイント一覧

ベース URL: `http://localhost:8080/api/v1`

### 認証

| メソッド | パス | 説明 |
|---|---|---|
| POST | `/auth/signup` | ユーザー登録 |
| POST | `/auth/login` | ログイン |
| GET | `/auth/me` | 自分の情報取得（JWT実装後に認証必須） |

> JWT 認証は未実装です。実装後に認証ミドルウェアを `router/router.go` に追加してください。

### ロケーション

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | `/locations` | 不要 | 共有ロケーション一覧 |
| GET | `/locations/nearby?lat=&lng=&radius_m=` | 不要 | 近くのロケーション取得 |
| GET | `/locations/:id/cards` | 不要 | ロケーションのカード一覧 |

### カード

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | `/cards/daily` | 不要 | 日常カード一覧 |
| POST | `/user/cards` | 要 | カード作成（multipart/form-data） |
| POST | `/user/locations/:id/cards` | 要 | カードをロケーションに追加 |
| DELETE | `/user/locations/:id/cards/:card_id` | 要 | カードをロケーションから削除 |
| PUT | `/user/locations/:id/cards/reorder` | 要 | カードの並び替え |

### ユーザーロケーション

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | `/user/locations` | 要 | 自分のロケーション一覧 |
| POST | `/user/locations` | 要 | ロケーション作成 |
| PUT | `/user/locations/:id` | 要 | ロケーション更新 |
| DELETE | `/user/locations/:id` | 要 | ロケーション削除 |
| GET | `/user/locations/:id/cards` | 要 | ロケーションのカード一覧 |

### AI

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| POST | `/ai/recommend` | 不要 | 文章候補生成（Gemini） |

---

## APIテスト例

### ユーザー登録

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","display_name":"テストユーザー"}'
```

### ログイン

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### 日常カード取得

```bash
curl -s http://localhost:8080/api/v1/cards/daily
```

### ユーザーロケーション作成（JWT実装後は Authorization ヘッダーが必要）

```bash
curl -s -X POST http://localhost:8080/api/v1/user/locations \
  -H "Content-Type: application/json" \
  -d '{"name":"自宅","latitude":35.6762,"longitude":139.6503,"radius_m":100}'
```

### AI 文章候補生成

```bash
curl -s -X POST http://localhost:8080/api/v1/ai/recommend \
  -H "Content-Type: application/json" \
  -d '{"words":["水","ください"],"location_name":"病院"}'
```

---

## コンテナ停止・削除

```bash
# 停止
sudo docker-compose down

# 停止 + ボリューム削除（DBデータも消えます）
sudo docker-compose down -v
```

---

## ディレクトリ構成

```
backend/
├── cmd/
│   ├── server/        - サーバーエントリーポイント
│   └── migrate/       - マイグレーション CLI
├── internal/
│   ├── config/        - 環境変数読み込み
│   ├── db/            - MySQL 接続
│   ├── handler/       - Gin ハンドラ
│   ├── middleware/    - CORS ミドルウェア
│   ├── model/         - リクエスト・レスポンス・DB モデル
│   ├── router/        - ルーティング設定
│   ├── service/       - ビジネスロジック
│   └── storage/       - Supabase Storage クライアント
└── migrations/        - MySQL DDL・シード SQL
```
