package record

// 單一記帳紀錄
type Record struct {
	ID           int64   `json:"id"`
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	CategoryIcon string  `json:"categoryIcon"`
	Tag          string  `json:"tag"`
	Note         string  `json:"note"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	CreatedAt    string  `json:"createdAt"`
	LineUserID   string  `json:"lineUserId"`
	// 查詢時由白名單帶出的記帳者顯示名稱，寫入時忽略。
	RecorderName string `json:"recorderName"`
}

// 支出分類
type SpendCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// LINE 用戶白名單
type LineUserWhitelist struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	CreatedAt   string `json:"createdAt"`
}
