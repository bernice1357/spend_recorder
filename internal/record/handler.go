package record

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// 紀錄查詢共用的欄位與 JOIN，順序需與 scanRecord 一致；記帳者名稱來自白名單。
const recordSelect = `SELECT r.id, r.category_id, COALESCE(c.name, ''), COALESCE(c.icon, ''), r.tag, r.note,
	 r.amount, r.date, COALESCE(r.created_at, ''), r.line_user_id, COALESCE(w.display_name, '')
	 FROM records r
	 LEFT JOIN spendCategories c ON r.category_id = c.id
	 LEFT JOIN lineUserWhitelist w ON r.line_user_id = w.user_id`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanRecord(row rowScanner) (Record, error) {
	var r Record
	err := row.Scan(&r.ID, &r.CategoryID, &r.CategoryName, &r.CategoryIcon, &r.Tag, &r.Note,
		&r.Amount, &r.Date, &r.CreatedAt, &r.LineUserID, &r.RecorderName)
	return r, err
}

// POST /api/records
func (h *Handler) Create(c *gin.Context) {
	var rec Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		log.Printf("[record] 解析新增記帳參數失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec.LineUserID = strings.TrimSpace(rec.LineUserID)
	if rec.LineUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lineUserId is required"})
		return
	}
	var allowed bool
	if err := h.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM lineUserWhitelist WHERE user_id = ?)`, rec.LineUserID,
	).Scan(&allowed); err != nil {
		log.Printf("[record] 新增記帳時查詢白名單失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate line user"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "line user is not authorized"})
		return
	}

	rec.CreatedAt = time.Now().Format(time.RFC3339)

	result, err := h.db.Exec(
		`INSERT INTO records (category_id, tag, note, amount, date, created_at, line_user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rec.CategoryID, rec.Tag, rec.Note, rec.Amount, rec.Date, rec.CreatedAt, rec.LineUserID,
	)
	if err != nil {
		log.Printf("[record] 新增記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("[record] 已新增紀錄但取得 ID 失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "record saved but failed to retrieve id"})
		return
	}
	out, err := scanRecord(h.db.QueryRow(recordSelect+` WHERE r.id=?`, id))
	if err != nil {
		log.Printf("[record] 讀取新增後的記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

// DELETE /api/records/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("[record] 解析刪除記帳 ID失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.db.Exec(`DELETE FROM records WHERE id = ?`, id)
	if err != nil {
		log.Printf("[record] 刪除記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		log.Printf("[record] 已執行刪除但取得影響筆數失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve affected rows"})
		return
	}
	if n == 0 {
		log.Printf("[record] 刪除失敗: record_id=%d 不存在", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}
	c.Status(http.StatusOK)
}

// PUT /api/records/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("[record] 解析更新記帳 ID失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var rec Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		log.Printf("[record] 解析更新記帳參數失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.db.Exec(
		`UPDATE records SET category_id=?, tag=?, note=?, amount=?, date=? WHERE id=?`,
		rec.CategoryID, rec.Tag, rec.Note, rec.Amount, rec.Date, id,
	)
	if err != nil {
		log.Printf("[record] 更新記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		log.Printf("[record] 已執行更新但取得影響筆數失敗: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve affected rows"})
		return
	}
	if n == 0 {
		log.Printf("[record] 更新失敗: record_id=%d 不存在", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}
	updated, err := scanRecord(h.db.QueryRow(recordSelect+` WHERE r.id=?`, id))
	if err != nil {
		log.Printf("[record] 讀取更新後的記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// GET /api/spend-categories
func (h *Handler) ListSpendCategories(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name, icon FROM spendCategories`)
	if err != nil {
		log.Printf("[record] 查詢分類失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	cats := []SpendCategory{}
	for rows.Next() {
		var cat SpendCategory
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon); err != nil {
			log.Printf("[record] 解析分類資料失敗: method=%s path=%s err=%v",
				c.Request.Method, c.FullPath(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cats = append(cats, cat)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[record] 讀取分類結果失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

// POST /api/spend-categories
func (h *Handler) CreateSpendCategory(c *gin.Context) {
	var cat SpendCategory
	if err := c.ShouldBindJSON(&cat); err != nil {
		log.Printf("[record] 解析新增分類參數失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cat.ID == "" || cat.Name == "" {
		log.Printf("[record] 新增分類失敗: 缺少分類 ID 或名稱")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and name are required"})
		return
	}
	if cat.Icon == "" {
		cat.Icon = "🏷️"
	}

	_, err := h.db.Exec(
		`INSERT INTO spendCategories (id, name, icon) VALUES (?, ?, ?)`,
		cat.ID, cat.Name, cat.Icon,
	)
	if err != nil {
		log.Printf("[record] 新增分類失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// GET /api/records?date=YYYY-MM-DD 或 month=YYYY-MM
func (h *Handler) List(c *gin.Context) {
	var (
		rows *sql.Rows
		err  error
	)
	if date := c.Query("date"); date != "" {
		rows, err = h.db.Query(recordSelect+` WHERE r.date = ? ORDER BY r.created_at DESC`, date)
	} else if month := c.Query("month"); month != "" {
		start, parseErr := time.Parse("2006-01", month)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month, expected YYYY-MM"})
			return
		}
		rows, err = h.db.Query(recordSelect+` WHERE r.date >= ? AND r.date < ? ORDER BY r.created_at DESC`,
			start.Format("2006-01-02"), start.AddDate(0, 1, 0).Format("2006-01-02"))
	} else {
		rows, err = h.db.Query(recordSelect + ` ORDER BY r.created_at DESC`)
	}
	if err != nil {
		log.Printf("[record] 查詢記帳失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	records := []Record{}
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			log.Printf("[record] 解析記帳資料失敗: method=%s path=%s err=%v",
				c.Request.Method, c.FullPath(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[record] 讀取記帳結果失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

// POST /api/validate-line-userId
// 驗證通過且帶有 displayName 時，順便更新白名單上的顯示名稱，供紀錄標示記帳者。
func (h *Handler) ValidateLineUserID(c *gin.Context) {
	var req struct {
		UserID      string `json:"userId" binding:"required"`
		DisplayName string `json:"displayName" binding:"max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[record] 解析白名單驗證參數失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	var count int
	if err := h.db.QueryRow(
		`SELECT COUNT(*) FROM lineUserWhitelist WHERE user_id = ?`, req.UserID,
	).Scan(&count); err != nil {
		log.Printf("[record] 查詢白名單失敗: method=%s path=%s err=%v",
			c.Request.Method, c.FullPath(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if name := strings.TrimSpace(req.DisplayName); count > 0 && name != "" {
		if _, err := h.db.Exec(
			`UPDATE lineUserWhitelist SET display_name = ? WHERE user_id = ?`, name, req.UserID,
		); err != nil {
			// 名稱只用於顯示，更新失敗不影響登入。
			log.Printf("[record] 更新白名單顯示名稱失敗: %v", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"valid": count > 0})
}
