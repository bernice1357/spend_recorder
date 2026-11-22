package line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.line.me"

// Client 是 LINE Messaging API 的客戶端
type Client struct {
	token      string
	httpClient *http.Client
}

// 用 access token 建立 client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_ACCESS_TOKEN 為空")
	}
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Reply 用 replyToken 回覆使用者（5 分鐘內有效）
func (c *Client) Reply(replyToken string, messages ...Message) error {
	return c.post("/v2/bot/message/reply", map[string]interface{}{
		"replyToken": replyToken,
		"messages":   messages,
	})
}

// Push 主動發訊息給指定使用者
func (c *Client) Push(to string, messages ...Message) error {
	return c.post("/v2/bot/message/push", map[string]interface{}{
		"to":       to,
		"messages": messages,
	})
}

func (c *Client) post(path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("編碼 LINE API 請求失敗: path=%s: %w", path, err)
	}
	req, err := http.NewRequest(http.MethodPost, apiBase+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("建立 LINE API 請求失敗: path=%s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("LINE API 連線失敗: path=%s: %w", path, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("讀取 LINE API 回應失敗: path=%s status=%d: %w", path, resp.StatusCode, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("呼叫 LINE API 失敗: path=%s status=%d request_id=%s",
			path, resp.StatusCode, resp.Header.Get("x-line-request-id"))
	}
	return nil
}
