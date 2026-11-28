-- 2. 依第 1 步的結果，把 A_ID / B_ID 換成實際的 userId，名字換成要顯示的名稱
BEGIN;

UPDATE lineUserWhitelist SET display_name = 'A的名字' WHERE user_id = 'A_ID';
UPDATE lineUserWhitelist SET display_name = 'B的名字' WHERE user_id = 'B_ID';

-- 確認結果：兩列都要有名字
SELECT user_id, display_name FROM lineUserWhitelist;

COMMIT;
