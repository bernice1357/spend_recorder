-- 1. 找出白名單上每個 userId 記了些什麼，用來判斷 ID 是誰
.headers on
.mode column

-- 每個 userId 的概況
SELECT
  w.user_id,
  w.display_name,
  w.created_at  AS joined_at,
  COUNT(r.id)   AS record_count,
  MAX(r.date)   AS last_date
FROM lineUserWhitelist w
LEFT JOIN records r ON r.line_user_id = w.user_id
GROUP BY w.user_id
ORDER BY w.created_at;

-- 每個 userId 最近 5 筆紀錄（日期、分類、tag、備註、金額）
SELECT user_id, date, category, tag, note, amount
FROM (
  SELECT
    w.user_id,
    r.date,
    c.name AS category,
    r.tag,
    r.note,
    r.amount,
    ROW_NUMBER() OVER (PARTITION BY w.user_id ORDER BY r.date DESC, r.id DESC) AS rn
  FROM lineUserWhitelist w
  JOIN records r ON r.line_user_id = w.user_id
  LEFT JOIN spendCategories c ON c.id = r.category_id
)
WHERE rn <= 5
ORDER BY user_id, date DESC;
