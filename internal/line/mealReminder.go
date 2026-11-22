package line

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"cash_carole_record/internal/record"
)

// 啟動台灣時間每天 20 點的提醒，回傳停止並等待排程結束的函式。
func StartDailyMealReminder(db *sql.DB, client *Client) func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	location := time.FixedZone("Asia/Taipei", 8*60*60)

	go func() {
		defer close(done)
		for {
			now := time.Now().In(location)
			next := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, location)
			if !now.Before(next) {
				next = next.AddDate(0, 0, 1)
			}
			log.Printf("[meal-reminder] 下次檢查: time=%s", next.Format(time.RFC3339))
			timer := time.NewTimer(time.Until(next))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				sendMealReminder(ctx, db, client, next.Format("2006-01-02"))
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

// 缺少任一餐時，將缺少的餐別私訊給每位白名單使用者。
func sendMealReminder(ctx context.Context, db *sql.DB, client *Client, date string) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	missing, err := record.MissingMeals(queryCtx, db, date)
	if err != nil {
		log.Printf("[meal-reminder] 檢查失敗: date=%s err=%v", date, err)
		return
	}
	if len(missing) == 0 {
		log.Printf("[meal-reminder] 三餐已齊全，不發送提醒: date=%s", date)
		return
	}

	rows, err := db.QueryContext(queryCtx, `SELECT user_id FROM lineUserWhitelist ORDER BY user_id`)
	if err != nil {
		log.Printf("[meal-reminder] 讀取白名單失敗: date=%s err=%v", date, err)
		return
	}
	defer rows.Close()
	var recipients []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			log.Printf("[meal-reminder] 解析白名單失敗: date=%s err=%v", date, err)
			return
		}
		recipients = append(recipients, userID)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[meal-reminder] 讀取白名單結果失敗: date=%s err=%v", date, err)
		return
	}
	if err := rows.Close(); err != nil {
		log.Printf("[meal-reminder] 關閉白名單查詢失敗: date=%s err=%v", date, err)
		return
	}
	cancel()

	message := NewText("今天（" + date + "）還沒記錄：" + strings.Join(missing, "、") + "，記得記帳。")
	sent := 0
	for index, userID := range recipients {
		if ctx.Err() != nil {
			return
		}
		if err := client.Push(userID, message); err != nil {
			log.Printf("[meal-reminder] 私訊失敗: date=%s recipient_index=%d err=%v", date, index, err)
			continue
		}
		sent++
	}
	log.Printf("[meal-reminder] 提醒完成: date=%s missing=%s recipients=%d sent=%d",
		date, strings.Join(missing, "、"), len(recipients), sent)
}
