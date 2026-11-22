package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	ChannelSecret      string
	ChannelAccessToken string
	LiffID             string
	GeminiAPIKey       string
	GeminiModel        string
}

// 從環境變數載入設定
func Load() *Config {
	// 載入 .env（若存在），不會覆蓋已設定的環境變數
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		// 解析錯誤可能包含環境變數內容，不輸出錯誤原文。
		log.Printf("[config] 載入 .env 失敗，請檢查檔案權限與格式: error_type=%T", err)
	}
	return &Config{
		Port:               getEnv("PORT", "8080"),
		ChannelSecret:      os.Getenv("LINE_CHANNEL_SECRET"),
		ChannelAccessToken: os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		LiffID:             os.Getenv("LIFF_ID"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-3.5-flash-lite"),
	}
}

func (c *Config) Validate() error {
	if c.ChannelSecret == "" {
		return fmt.Errorf("LINE_CHANNEL_SECRET 為必填（用於驗證 webhook signature）")
	}
	if c.ChannelAccessToken == "" {
		return fmt.Errorf("LINE_CHANNEL_ACCESS_TOKEN 為必填")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
