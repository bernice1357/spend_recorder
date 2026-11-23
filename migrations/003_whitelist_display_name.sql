-- 適用於尚未有 display_name 欄位的既有資料庫；執行前先備份，本檔不可重複執行。
-- display_name：白名單使用者的 LINE 顯示名稱，下次從 LINE 登入時自動更新，用於標示記帳者。
BEGIN IMMEDIATE;

ALTER TABLE lineUserWhitelist ADD COLUMN display_name TEXT NOT NULL DEFAULT '';

COMMIT;
