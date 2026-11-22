package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"cash_carole_record/internal/config"
	"cash_carole_record/internal/line"
	"cash_carole_record/internal/logging"
	"cash_carole_record/internal/record"
)

func main() {
	appLog, err := logging.NewDailyWriter("logs", "app")
	if err != nil {
		log.Fatalf("[log] 開啟一般日誌失敗: %v", err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, appLog))
	defer func() {
		log.SetOutput(os.Stderr)
		if err := appLog.Close(); err != nil {
			log.Printf("[log] 關閉一般日誌失敗: %v", err)
		}
	}()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[config] %v", err)
	}

	ginLog, err := logging.NewDailyWriter("logs", "gin")
	if err != nil {
		log.Fatalf("[log] 開啟 Gin 日誌失敗: %v", err)
	}
	defer func() {
		if err := ginLog.Close(); err != nil {
			log.Printf("[log] 關閉 Gin 日誌失敗: %v", err)
		}
	}()
	webhookLog, err := logging.NewDailyWriter("logs", "webhook")
	if err != nil {
		log.Fatalf("[log] 開啟 webhook 日誌失敗: %v", err)
	}
	defer func() {
		if err := webhookLog.Close(); err != nil {
			log.Printf("[log] 關閉 webhook 日誌失敗: %v", err)
		}
	}()
	webhookLogger := log.New(webhookLog, "", log.LstdFlags)
	gin.DisableConsoleColor()
	gin.DefaultWriter = ginLog
	gin.DefaultErrorWriter = ginLog

	client, err := line.NewClient(cfg.ChannelAccessToken)
	if err != nil {
		log.Fatalf("[line] 初始化 client 失敗: %v", err)
	}

	db, err := record.InitDB("data/records.db")
	if err != nil {
		log.Fatalf("[record] 初始化 SQLite 失敗: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("[record] 關閉 SQLite 失敗: %v", err)
		}
	}()
	recHandler := record.NewHandler(db)
	analyzer := record.NewAnalyzer(db, cfg.GeminiAPIKey, cfg.GeminiModel)
	stopMealReminder := line.StartDailyMealReminder(db, client)
	defer stopMealReminder()

	r := gin.Default()

	r.POST("/webhook", line.WebhookHandler(client, cfg.ChannelSecret, db, analyzer, webhookLogger))
	r.POST("/api/records", recHandler.Create)
	r.PUT("/api/records/:id", recHandler.Update)
	r.DELETE("/api/records/:id", recHandler.Delete)
	r.GET("/api/records", recHandler.List)
	r.GET("/api/spend-categories", recHandler.ListSpendCategories)
	r.POST("/api/spend-categories", recHandler.CreateSpendCategory)
	r.POST("/api/validate-line-userId", recHandler.ValidateLineUserID)
	r.POST("/api/log", func(c *gin.Context) {
		var req struct {
			Msg   string `json:"msg" binding:"required,max=4096"`
			Level string `json:"level" binding:"omitempty,oneof=info warn error"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[frontend-log] 解析日誌請求失敗: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if req.Level == "" {
			req.Level = "info"
		}
		log.Printf("[frontend-log] level=%s message=%q", req.Level, req.Msg)
		c.JSON(200, gin.H{"ok": true})
	})
	r.GET("/api/liff-config", func(c *gin.Context) {
		c.JSON(200, gin.H{"liffId": cfg.LiffID})
	})

	// 未匹配到路由時，fallback 到靜態檔案
	fileServer := http.FileServer(http.Dir("internal/web-frontend"))
	r.NoRoute(func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	addr := ":" + cfg.Port
	log.Printf("[main] 服務啟動，監聽 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[main] %v", err)
	}
}
