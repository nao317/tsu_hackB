#!/bin/bash
# =============================================
# API エンドポイント テストスクリプト
# 対象: 認証不要エンドポイント
# 使い方: bash test_api.sh
# =============================================

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0

# --- ユーティリティ ---
check() {
  local name="$1"
  local expected_status="$2"
  local actual_status="$3"
  local body="$4"
  local extra_check="$5"  # bodyに含まれるべき文字列（任意）

  local ok=true

  if [ "$actual_status" != "$expected_status" ]; then
    ok=false
  fi

  if [ -n "$extra_check" ] && ! echo "$body" | grep -q "$extra_check"; then
    ok=false
  fi

  if $ok; then
    echo "  ✅ PASS  $name"
    PASS=$((PASS + 1))
  else
    echo "  ❌ FAIL  $name"
    echo "       期待ステータス: $expected_status / 実際: $actual_status"
    if [ -n "$extra_check" ] && ! echo "$body" | grep -q "$extra_check"; then
      echo "       レスポンスに '$extra_check' が含まれていません"
      echo "       レスポンス: $body"
    fi
    FAIL=$((FAIL + 1))
  fi
}

echo ""
echo "============================================"
echo " API テスト開始"
echo " Base URL: $BASE"
echo "============================================"

# =============================================
# 1. ヘルスチェック
# =============================================
echo ""
echo "[ 1. ヘルスチェック ]"

RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" http://localhost:8080/health)
BODY=$(cat /tmp/body.txt)
check "GET /health → 200" "200" "$RES" "$BODY" "ok"

# =============================================
# 2. ロケーション
# =============================================
echo ""
echo "[ 2. ロケーション ]"

RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" "$BASE/locations")
BODY=$(cat /tmp/body.txt)
check "GET /locations → 200" "200" "$RES" "$BODY"
check "GET /locations → コンビニを含む" "200" "$RES" "$BODY" "コンビニ"
check "GET /locations → 病院を含む" "200" "$RES" "$BODY" "病院"
check "GET /locations → カフェを含む" "200" "$RES" "$BODY" "カフェ"

# nearby（seedデータに緯度経度がないため空配列 or null が正常）
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" \
  "$BASE/locations/nearby?lat=35.6762&lng=139.6503&radius_m=500")
BODY=$(cat /tmp/body.txt)
check "GET /locations/nearby → 200" "200" "$RES" "$BODY"

# nearby パラメータ欠け → 400
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" "$BASE/locations/nearby")
BODY=$(cat /tmp/body.txt)
check "GET /locations/nearby (パラメータなし) → 400" "400" "$RES" "$BODY"

# locations/:id/cards（seedデータにカード紐付けなしのため空 or null が正常）
LOC_ID=$(curl -s "$BASE/locations" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])" 2>/dev/null)
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" "$BASE/locations/$LOC_ID/cards")
BODY=$(cat /tmp/body.txt)
check "GET /locations/:id/cards → 200" "200" "$RES" "$BODY"

# 存在しないID → 200（空配列）
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" "$BASE/locations/00000000-0000-0000-0000-000000000000/cards")
BODY=$(cat /tmp/body.txt)
check "GET /locations/不正ID/cards → 200(空)" "200" "$RES" "$BODY"

# =============================================
# 3. 日常カード
# =============================================
echo ""
echo "[ 3. 日常カード ]"

RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" "$BASE/cards/daily")
BODY=$(cat /tmp/body.txt)
check "GET /cards/daily → 200" "200" "$RES" "$BODY"
check "GET /cards/daily → こんにちは を含む" "200" "$RES" "$BODY" "こんにちは"
check "GET /cards/daily → ありがとう を含む" "200" "$RES" "$BODY" "ありがとう"
check "GET /cards/daily → 7件含む" "200" "$RES" \
  "$(echo "$BODY" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null)" "7"

# =============================================
# 4. 認証
# =============================================
echo ""
echo "[ 4. 認証 ]"

# テスト用ユニークメール
TEST_EMAIL="testuser_$$@example.com"

# signup 正常
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"password123\",\"display_name\":\"テストユーザー\"}")
BODY=$(cat /tmp/body.txt)
check "POST /auth/signup → 201" "201" "$RES" "$BODY"
check "POST /auth/signup → id を含む" "201" "$RES" "$BODY" "\"id\""
check "POST /auth/signup → email を含む" "201" "$RES" "$BODY" "$TEST_EMAIL"

# signup 重複メール → 409
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"password123\",\"display_name\":\"重複\"}")
BODY=$(cat /tmp/body.txt)
check "POST /auth/signup (重複) → 409" "409" "$RES" "$BODY" "EMAIL_ALREADY_EXISTS"

# signup バリデーションエラー（emailが不正）→ 400
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{"email":"not-an-email","password":"password123","display_name":"テスト"}')
BODY=$(cat /tmp/body.txt)
check "POST /auth/signup (不正email) → 400" "400" "$RES" "$BODY" "VALIDATION_ERROR"

# signup バリデーションエラー（password短すぎ）→ 400
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{"email":"short@example.com","password":"abc","display_name":"テスト"}')
BODY=$(cat /tmp/body.txt)
check "POST /auth/signup (password短すぎ) → 400" "400" "$RES" "$BODY" "VALIDATION_ERROR"

# login 正常
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"password123\"}")
BODY=$(cat /tmp/body.txt)
check "POST /auth/login → 200" "200" "$RES" "$BODY"
check "POST /auth/login → id を含む" "200" "$RES" "$BODY" "\"id\""

# login 誤パスワード → 401
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"wrongpassword\"}")
BODY=$(cat /tmp/body.txt)
check "POST /auth/login (誤パスワード) → 401" "401" "$RES" "$BODY" "INVALID_CREDENTIALS"

# login 存在しないメール → 401
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"notexist@example.com","password":"password123"}')
BODY=$(cat /tmp/body.txt)
check "POST /auth/login (存在しないメール) → 401" "401" "$RES" "$BODY" "INVALID_CREDENTIALS"

# =============================================
# 5. AI
# =============================================
echo ""
echo "[ 5. AI レコメンド ]"

RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/ai/recommend" \
  -H "Content-Type: application/json" \
  -d '{"words":["水","ください"],"location_name":"病院"}')
BODY=$(cat /tmp/body.txt)
check "POST /ai/recommend → 200" "200" "$RES" "$BODY"
check "POST /ai/recommend → latency_ms を含む" "200" "$RES" "$BODY" "latency_ms"

# wordが空 → 400
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/ai/recommend" \
  -H "Content-Type: application/json" \
  -d '{"words":[]}')
BODY=$(cat /tmp/body.txt)
check "POST /ai/recommend (words空) → 400" "400" "$RES" "$BODY" "VALIDATION_ERROR"

# words フィールドなし → 400
RES=$(curl -s -o /tmp/body.txt -w "%{http_code}" -X POST "$BASE/ai/recommend" \
  -H "Content-Type: application/json" \
  -d '{}')
BODY=$(cat /tmp/body.txt)
check "POST /ai/recommend (wordsなし) → 400" "400" "$RES" "$BODY" "VALIDATION_ERROR"

# =============================================
# 結果サマリー
# =============================================
TOTAL=$((PASS + FAIL))
echo ""
echo "============================================"
echo " テスト結果: $PASS / $TOTAL 件 PASS"
if [ $FAIL -gt 0 ]; then
  echo " ❌ FAIL: $FAIL 件"
else
  echo " 全テスト PASS ✅"
fi
echo "============================================"
echo ""

[ $FAIL -eq 0 ]
