package line

// WebhookRequest 是 LINE 送來的請求 body
type WebhookRequest struct {
	Destination   string  `json:"destination"`
	WebhookHeader *Header `json:"webhookHeader,omitempty"`
	Events        []Event `json:"events"`
}

// Header 是 LINE API 版本資訊
type Header struct {
	LineApiVersion string `json:"lineApiVersion,omitempty"`
}

// Event 是單一事件
type Event struct {
	Type       string          `json:"type"`
	Mode       string          `json:"mode,omitempty"`
	CreateTime int64           `json:"createTime,omitempty"`
	Source     Source          `json:"source"`
	ReplyToken string          `json:"replyToken,omitempty"`
	Message    *EventMessage   `json:"message,omitempty"`
	Follow     *Follow         `json:"follow,omitempty"`
	Unfollow   *Unfollow       `json:"unfollow,omitempty"`
	Join       *Join           `json:"join,omitempty"`
	Leave      *Leave          `json:"leave,omitempty"`
	Postback   *Postback       `json:"postback,omitempty"`
}

// Source 是事件來源（user / group / room）
type Source struct {
	Type    string  `json:"type"`
	UserID  *string `json:"userId,omitempty"`
	GroupID *string `json:"groupId,omitempty"`
	RoomID  *string `json:"roomId,omitempty"`
}

// EventMessage 是使用者傳來的訊息
type EventMessage struct {
	ID   string  `json:"id"`
	Type string  `json:"type"` // text / image / video / audio / file / location / sticker
	Text *string `json:"text,omitempty"`
}

// Message 是機器人要送的訊息
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewText 建立文字訊息
func NewText(text string) Message {
	return Message{Type: "text", Text: text}
}

// Follow 是關注事件
type Follow struct {
	User *User `json:"user,omitempty"`
	Referral *struct {
		Parameter map[string]string `json:"parameter,omitempty"`
	} `json:"referral,omitempty"`
}

// User 是使用者資訊
type User struct {
	UserID      string  `json:"userId,omitempty"`
	Name        string  `json:"name,omitempty"`
	DisplayName string  `json:"displayName,omitempty"`
	PictureURL  *string `json:"pictureUrl,omitempty"`
}

// Unfollow 是取消關注事件
type Unfollow struct{}

// Join 是加入群組/房間事件
type Join struct{}

// Leave 是離開群組/房間事件
type Leave struct{}

// Postback 是 postback 事件
type Postback struct {
	Data       string            `json:"data,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}