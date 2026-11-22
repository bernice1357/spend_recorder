-- 僅適用於尚未有分類外鍵的 records，執行前須備份並確認沒有孤立紀錄。
-- 原表若有額外欄位、索引、trigger 或被其他表參照，須先調整遷移方案。
PRAGMA foreign_keys = ON;
BEGIN IMMEDIATE;

CREATE TEMP TABLE records_sequence_before_category_fk (
  seq INTEGER NOT NULL
);
INSERT INTO records_sequence_before_category_fk (seq)
SELECT COALESCE(MAX(seq), 0) FROM sqlite_sequence WHERE name = 'records';

CREATE TABLE records_with_category_fk (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  category_id TEXT    NOT NULL,
  tag         TEXT    NOT NULL DEFAULT '',
  note        TEXT    NOT NULL DEFAULT '',
  amount      REAL    NOT NULL,
  date        TEXT    NOT NULL,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (category_id) REFERENCES spendCategories(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
);

INSERT INTO records_with_category_fk (id, category_id, tag, note, amount, date, created_at)
SELECT id, category_id, tag, note, amount, date, created_at FROM records;

DROP TABLE records;
ALTER TABLE records_with_category_fk RENAME TO records;

-- 即使曾刪除最大 ID，或原表目前沒有資料，也保留原本的遞增序號。
INSERT INTO sqlite_sequence (name, seq)
SELECT 'records', seq FROM records_sequence_before_category_fk
WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = 'records');
UPDATE sqlite_sequence
SET seq = MAX(seq, (SELECT seq FROM records_sequence_before_category_fk))
WHERE name = 'records';

CREATE INDEX idx_records_category_id ON records(category_id);
DROP TABLE temp.records_sequence_before_category_fk;

PRAGMA foreign_key_check;
COMMIT;
