package line

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"time"
	"unicode/utf16"

	"cash_carole_record/internal/record"

	"github.com/gin-gonic/gin"
)

// 回傳處理 LINE webhook 的 handler
func WebhookHandler(
	client *Client, channelSecret string, db *sql.DB, analyzer *record.Analyzer, logger *log.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf("[webhook] 請求處理 panic: error_type=%T duration=%s", recovered, time.Since(started))
				panic(recovered)
			}
			logger.Printf("[webhook] 請求處理結束: status=%d duration=%s", c.Writer.Status(), time.Since(started))
		}()
		rawBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Printf("[line] 讀取 webhook 請求失敗: %v", err)
			c.AbortWithStatus(400)
			return
		}

		// 驗證 X-Line-Signature，不符視為非 LINE 送來的請求
		if !VerifySignature(rawBody, c.GetHeader("X-Line-Signature"), channelSecret) {
			logger.Printf("[line] webhook 簽章驗證失敗")
			c.AbortWithStatus(401)
			return
		}

		var req WebhookRequest
		if err := json.Unmarshal(rawBody, &req); err != nil {
			logger.Printf("[line] 解析 webhook JSON 失敗: %v", err)
			c.AbortWithStatus(400)
			return
		}

		logger.Printf("[webhook] 收到事件: count=%d", len(req.Events))
		for _, event := range req.Events {
			handleEvent(c.Request.Context(), client, db, analyzer, event, logger)
		}

		// LINE 要求回 2xx，否則會持續重試
		c.Status(200)
	}
}

// 依事件類型處理
func handleEvent(
	ctx context.Context, client *Client, db *sql.DB, analyzer *record.Analyzer, event Event, logger *log.Logger,
) {
	logger.Printf("[webhook] 開始處理事件: event=%s", event.Type)
	// 白名單檢查：userId 不在白名單則跳過
	if event.Source.UserID == nil || event.ReplyToken == "" {
		logger.Printf("[line] 略過無法回覆的事件: event=%s 缺少 userId 或 replyToken", event.Type)
		return
	}
	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM lineUserWhitelist WHERE user_id = ?`, *event.Source.UserID,
	).Scan(&count); err != nil {
		logger.Printf("[line] 白名單查詢失敗: %v", err)
		return
	}
	if count == 0 {
		logger.Printf("[line] 略過未授權的事件: event=%s", event.Type)
		return
	}

	var messages []Message

	switch event.Type {
	case "message":
		if event.Message == nil || event.Message.Type != "text" || event.Message.Text == nil {
			logger.Printf("[line] 略過訊息: 非文字訊息或缺少文字內容")
			return
		}
		// 分析不隨 webhook 斷線取消，仍限制最多執行 40 秒。
		analysisCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 40*time.Second)
		answer, err := analyzer.Answer(analysisCtx, *event.Message.Text)
		cancel()
		if err != nil {
			logger.Printf("[line] 記帳分析失敗: %v", err)
		}
		// LINE 文字長度以 UTF-16 計算，預留截斷提示的空間。
		units := utf16.Encode([]rune(answer))
		if len(units) > 4900 {
			end := 4900
			if units[end-1] >= 0xD800 && units[end-1] <= 0xDBFF {
				end--
			}
			answer = string(utf16.Decode(units[:end])) + "\n內容較長，請縮小查詢範圍。"
		}
		messages = []Message{NewText(answer)}

	case "postback":
		if event.Postback == nil {
			logger.Printf("[line] postback 事件缺少資料")
			return
		}
		messages = []Message{NewText("postback: " + event.Postback.Data)}

	default:
		logger.Printf("收到 webhook: %s", event.Type)
		return
	}

	if event.ReplyToken == "" {
		return
	}
	if err := client.Reply(event.ReplyToken, messages...); err != nil {
		logger.Printf("[line] 回覆失敗: event=%s err=%v", event.Type, err)
		return
	}
	logger.Printf("[line] 回覆成功: event=%s", event.Type)
}
