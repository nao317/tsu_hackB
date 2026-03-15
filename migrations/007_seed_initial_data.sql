-- 共有ロケーションのシード
INSERT INTO locations (id, name, radius_m, is_default) VALUES
  (UUID(), 'コンビニ', 100, 1),
  (UUID(), '病院',     200, 1),
  (UUID(), 'カフェ',   100, 1);

-- 日常カードのシード
INSERT INTO cards (id, label, emoji, is_daily) VALUES
  (UUID(), 'こんにちは', '👋', 1),
  (UUID(), 'ありがとう', '🙏', 1),
  (UUID(), 'すみません', '🙇', 1),
  (UUID(), 'はい',       '✅', 1),
  (UUID(), 'いいえ',     '❌', 1),
  (UUID(), 'おねがい',   '🙏', 1),
  (UUID(), 'わかった',   '👌', 1);