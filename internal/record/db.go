package record

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// 三餐分類，名稱需與 MissingMeals 檢查的餐別一致。
var mealCategories = []SpendCategory{
	{"breakfast", "早餐", "🍳"},
	{"lunch", "午餐", "🍱"},
	{"dinner", "晚餐", "🍲"},
}

// 初始化 SQLite 連線並建表
func InitDB(path string) (*sql.DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("取得資料庫絕對路徑失敗: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, fmt.Errorf("建立資料庫目錄失敗: %w", err)
	}
	params := url.Values{}
	// DSN 設定會套用至連線池的每個新連線。
	params.Set("_pragma", "foreign_keys(1)")
	slashPath := filepath.ToSlash(absPath)
	if !strings.HasPrefix(slashPath, "/") {
		// Windows 的 C:/... 需寫成 file:///C:/...，否則磁碟代號會被當成 URI authority。
		slashPath = "/" + slashPath
	}
	dsn := url.URL{Scheme: "file", Path: slashPath, RawQuery: params.Encode()}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("開啟 SQLite 失敗: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS records (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id TEXT    NOT NULL,
		tag         TEXT    NOT NULL DEFAULT '',
		note        TEXT    NOT NULL DEFAULT '',
		amount      REAL    NOT NULL,
		date        TEXT    NOT NULL,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		line_user_id TEXT   NOT NULL DEFAULT '',
		FOREIGN KEY (category_id) REFERENCES spendCategories(id)
			ON UPDATE CASCADE ON DELETE RESTRICT
	);
	CREATE TABLE IF NOT EXISTS spendCategories (
		id   TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		icon TEXT NOT NULL DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS lineUserWhitelist (
		user_id      TEXT PRIMARY KEY,
		display_name TEXT NOT NULL DEFAULT '',
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_records_category_id ON records(category_id);`

	if _, err := db.Exec(schema); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("[record] 建表失敗後關閉連線失敗: %v", closeErr)
		}
		return nil, fmt.Errorf("初始化資料表失敗: %w", err)
	}

	// seed 預設分類（只在表是空的時插入）
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM spendCategories`).Scan(&count); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("[record] 分類查詢失敗後關閉連線失敗: %v", closeErr)
		}
		return nil, fmt.Errorf("查詢預設分類筆數失敗: %w", err)
	}
	if count == 0 {
		defaults := []SpendCategory{
			{"snacks", "零食", "🍪"},
			{"for-anna", "給阿拿的錢", "💸"},
			{"daily", "日用品", "🧴"},
			{"drinks", "飲料", "🥤"},
			{"income", "收入", "💰"},
			{"transport", "交通", "🚌"},
			{"fun", "娛樂", "🎮"},
			{"shopping", "購物", "🛍️"},
			{"medical", "醫療", "💊"},
			{"education", "教育", "📚"},
			{"home", "居家", "🏠"},
			{"pet", "寵物", "🐱"},
			{"other", "其他", "📦"},
		}
		for _, c := range defaults {
			if _, err := db.Exec(
				`INSERT INTO spendCategories (id, name, icon) VALUES (?, ?, ?)`, c.ID, c.Name, c.Icon,
			); err != nil {
				if closeErr := db.Close(); closeErr != nil {
					log.Printf("[record] 分類寫入失敗後關閉連線失敗: %v", closeErr)
				}
				return nil, fmt.Errorf("寫入預設分類失敗: category_id=%s: %w", c.ID, err)
			}
			count++
		}
	}

	// 三餐提醒依同名分類判斷；既有資料庫缺少時補建，已有同名或同 ID 分類則略過。
	for _, c := range mealCategories {
		res, err := db.Exec(
			`INSERT OR IGNORE INTO spendCategories (id, name, icon)
			 SELECT ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM spendCategories WHERE name = ?)`,
			c.ID, c.Name, c.Icon, c.Name,
		)
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("[record] 三餐分類寫入失敗後關閉連線失敗: %v", closeErr)
			}
			return nil, fmt.Errorf("寫入三餐分類失敗: category_id=%s: %w", c.ID, err)
		}
		if n, err := res.RowsAffected(); err == nil && n > 0 {
			log.Printf("[record] 已補建三餐分類: category_id=%s", c.ID)
			count++
		}
	}

	log.Printf("[record] SQLite 已就緒: path=%s categories=%d", absPath, count)
	return db, nil
}
