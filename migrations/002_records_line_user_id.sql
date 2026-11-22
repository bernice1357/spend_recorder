-- 僅適用於尚未有 line_user_id 的既有資料庫；執行前先備份。
-- 若需分類外鍵遷移，先執行 001，再執行本檔；本檔不可重複執行。
-- 既有紀錄的作者依帳本擁有者指定，不使用此值作為之後新增紀錄的預設作者。
BEGIN IMMEDIATE;

ALTER TABLE records ADD COLUMN line_user_id TEXT NOT NULL DEFAULT '';

UPDATE records
SET line_user_id = 'Ud4af9c2ce7cca29234bfeb60b9da4198';

COMMIT;
