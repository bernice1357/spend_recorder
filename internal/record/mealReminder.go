package record

import (
	"context"
	"database/sql"
	"fmt"
)

// 依分類名稱檢查指定日期的共用帳本，回傳尚未記錄的餐別。
func MissingMeals(ctx context.Context, db *sql.DB, date string) ([]string, error) {
	meals := []string{"早餐", "午餐", "晚餐"}
	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT c.name
		 FROM records r JOIN spendCategories c ON c.id = r.category_id
		 WHERE r.date = ? AND c.name IN (?, ?, ?)`,
		date, meals[0], meals[1], meals[2],
	)
	if err != nil {
		return nil, fmt.Errorf("查詢三餐紀錄失敗: %w", err)
	}
	defer rows.Close()

	recorded := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("解析三餐紀錄失敗: %w", err)
		}
		recorded[name] = true
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
