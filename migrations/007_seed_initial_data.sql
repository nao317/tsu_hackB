-- 共有ロケーションのシード
INSERT INTO locations (
  id,
  name,
  description,
  latitude,
  longitude,
  radius_m,
  is_default,
  created_at
) VALUES
  (UUID(), 'コンビニ', '日用品や飲み物を購入する場所', 33.3373176, 131.4673964, 100, 1, NOW()),
  (UUID(), '大学', '大学のキャンパス内', 33.3370848, 131.4683736, 100, 1, NOW()),
  (UUID(), 'カフェ',   '飲み物や軽食を注文する場所',   33.337085, 139.701636, 100, 1, NOW());

-- 日常カードのシード
INSERT INTO cards (id, label, emoji, is_daily) VALUES
  (UUID(), '音声アプリで会話します。少しお待ちください', '🎤', 1),
  (UUID(), 'こんにちは', '👋', 1),
  (UUID(), 'すこしまってください', '⏳', 1),
  (UUID(), 'ありがとうございます', '🙏', 1);

-- ロケーション専用カードのシード
INSERT INTO cards (id, label, emoji, category, is_daily) VALUES
  (UUID(), 'レジ袋をください', '🛍️', 'location:convenience', 0),
  (UUID(), '温めてください', '🔥', 'location:convenience', 0),
  (UUID(), 'お会計をお願いします', '💳', 'location:convenience', 0),
  (UUID(), '受付はどこですか', '🏥', 'location:hospital', 0),
  (UUID(), '薬を受け取りたいです', '💊', 'location:hospital', 0),
  (UUID(), '診察をお願いします', '🩺', 'location:hospital', 0),
  (UUID(), 'おすすめをください', '☕', 'location:cafe', 0),
  (UUID(), '持ち帰りでお願いします', '🥤', 'location:cafe', 0),
  (UUID(), 'お水をください', '💧', 'location:cafe', 0);

-- 共有ロケーション × ロケーション専用カードの紐付けシード
INSERT INTO location_cards (id, location_id, card_id, sort_order)
SELECT
  UUID(),
  l.id,
  c.id,
  CASE
    WHEN l.name = 'コンビニ' AND c.label = 'レジ袋をください' THEN 0
    WHEN l.name = 'コンビニ' AND c.label = '温めてください' THEN 1
    WHEN l.name = 'コンビニ' AND c.label = 'お会計をお願いします' THEN 2
    WHEN l.name = '大学' AND c.label = '受付はどこですか' THEN 0
    WHEN l.name = '大学' AND c.label = '薬を受け取りたいです' THEN 1
    WHEN l.name = '大学' AND c.label = '診察をお願いします' THEN 2
    WHEN l.name = 'カフェ' AND c.label = 'おすすめをください' THEN 0
    WHEN l.name = 'カフェ' AND c.label = '持ち帰りでお願いします' THEN 1
    WHEN l.name = 'カフェ' AND c.label = 'お水をください' THEN 2
  END AS sort_order
FROM locations l
JOIN cards c ON (
  (l.name = 'コンビニ' AND c.category = 'location:convenience') OR
  (l.name = '大学' AND c.category = 'location:hospital') OR
  (l.name = 'カフェ' AND c.category = 'location:cafe')
)
WHERE l.name IN ('コンビニ', '大学', 'カフェ');
