package record

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// 三餐以同名分類記錄；舊資料若以 tag 標記餐別也視為已記錄。
var meals = []string{"早餐", "午餐", "晚餐"}

// splitTags 將以逗號分隔的 tag 欄位拆成個別標記。
func splitTags(tag string) []string {
	var tags []string
	for _, t := range strings.Split(tag, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// 檢查指定日期的共用帳本，回傳尚未記錄的餐別。
func MissingMeals(ctx context.Context, db *sql.DB, date string) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT COALESCE(c.name, ''), r.tag
		 FROM records r LEFT JOIN spendCategories c ON c.id = r.category_id
		 WHERE r.date = ?`, date)
	if err != nil {
		return nil, fmt.Errorf("查詢三餐紀錄失敗: %w", err)
	}
	defer rows.Close()

	recorded := make(map[string]bool)
	for rows.Next() {
		var name, tag string
		if err := rows.Scan(&name, &tag); err != nil {
			return nil, fmt.Errorf("解析三餐紀錄失敗: %w", err)
		}
		recorded[name] = true
		for _, t := range splitTags(tag) {
			recorded[t] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("讀取三餐紀錄失敗: %w", err)
	}

	var missing []string
	for _, meal := range meals {
		if !recorded[meal] {
			missing = append(missing, meal)
		}
	}
	return missing, nil
}
