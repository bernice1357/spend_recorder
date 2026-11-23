package record

import (
	"context"
	"errors"
	"strings"
	"time"
)

type analysisCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type analysisQuery struct {
	Label       string   `json:"label"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	CategoryIDs []string `json:"categoryIds"`
	Kind        string   `json:"kind"`
	Keyword     string   `json:"keyword"`
	Tag         string   `json:"tag"`
	GroupBy     string   `json:"groupBy"`
	Details     bool     `json:"details"`
	Order       string   `json:"order"`
	Limit       int      `json:"limit"`
}

type analysisTotals struct {
	Count        int64  `json:"count"`
	TotalCents   int64  `json:"totalCents"`
	IncomeCents  int64  `json:"incomeCents"`
	ExpenseCents int64  `json:"expenseCents"`
	BalanceCents int64  `json:"balanceCents"`
	MinCents     int64  `json:"minCents"`
	MaxCents     int64  `json:"maxCents"`
	FirstDate    string `json:"firstDate"`
	LastDate     string `json:"lastDate"`
}

type analysisGroup struct {
	Key string `json:"key"`
	analysisTotals
}

type analysisRecord struct {
	ID           int64  `json:"id"`
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Tag          string `json:"tag"`
	Note         string `json:"note"`
	AmountCents  int64  `json:"amountCents"`
	Date         string `json:"date"`
	CreatedAt    string `json:"createdAt"`
}

type analysisResult struct {
	Query            analysisQuery    `json:"query"`
	Summary          analysisTotals   `json:"summary"`
	Groups           []analysisGroup  `json:"groups,omitempty"`
	GroupsTruncated  bool             `json:"groupsTruncated"`
	Records          []analysisRecord `json:"records,omitempty"`
	RecordsTruncated bool             `json:"recordsTruncated"`
}

// 欄位與排序僅能使用程式定義的白名單，使用者條件一律綁定參數。
var analysisGroupColumns = map[string]string{
	"none":     "",
	"category": "r.category_id || ':' || COALESCE(c.name, r.category_id)",
	"day":      "r.date",
	"month":    "substr(r.date, 1, 7)",
	"tag":      "r.tag",
}

var analysisOrders = map[string]string{
	"date_desc":   "r.date DESC, r.id DESC",
	"amount_desc": "r.amount DESC, r.id DESC",
	"amount_asc":  "r.amount ASC, r.id DESC",
}

// 先將每筆金額轉為百分之一單位的整數，避免直接累加 REAL 的浮點誤差。
const analysisTotalColumns = `COUNT(*),
	COALESCE(SUM(CAST(ROUND(r.amount * 100) AS INTEGER)), 0),
	COALESCE(SUM(CASE WHEN r.category_id = 'income'
		THEN CAST(ROUND(r.amount * 100) AS INTEGER) ELSE 0 END), 0),
	COALESCE(SUM(CASE WHEN r.category_id <> 'income'
		THEN CAST(ROUND(r.amount * 100) AS INTEGER) ELSE 0 END), 0),
	COALESCE(MIN(CAST(ROUND(r.amount * 100) AS INTEGER)), 0),
	COALESCE(MAX(CAST(ROUND(r.amount * 100) AS INTEGER)), 0),
	COALESCE(MIN(r.date), ''), COALESCE(MAX(r.date), '')`

func (a *Analyzer) analysisCategories(ctx context.Context) ([]analysisCategory, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT id, name FROM spendCategories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories := make([]analysisCategory, 0)
	for rows.Next() {
		var category analysisCategory
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (q analysisQuery) validate(categories []analysisCategory) error {
	for _, date := range []string{q.StartDate, q.EndDate} {
		if date == "" {
			continue
		}
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil || parsed.Format("2006-01-02") != date {
			return errors.New("日期格式不正確")
		}
	}
	if q.StartDate != "" && q.EndDate != "" && q.StartDate > q.EndDate {
		return errors.New("日期範圍不正確")
	}
	if q.Kind != "all" && q.Kind != "income" && q.Kind != "expense" {
		return errors.New("收支類型不正確")
	}
	if _, ok := analysisGroupColumns[q.GroupBy]; !ok {
		return errors.New("分組條件不正確")
	}
	if _, ok := analysisOrders[q.Order]; !ok || q.Limit < 1 || q.Limit > 50 {
		return errors.New("明細條件不正確")
	}
	if len(q.CategoryIDs) > len(categories) || len(q.Keyword) > 2000 || len(q.Tag) > 2000 {
		return errors.New("篩選條件過長")
	}
	for _, id := range q.CategoryIDs {
		found := false
		for _, category := range categories {
			if category.ID == id {
				found = true
				break
			}
		}
		if !found {
			return errors.New("分類不存在")
		}
	}
	return nil
}

func (q analysisQuery) filter() (string, []interface{}) {
	conditions := []string{"1 = 1"}
	args := make([]interface{}, 0)
	if q.StartDate != "" {
		conditions = append(conditions, "r.date >= ?")
		args = append(args, q.StartDate)
	}
	if q.EndDate != "" {
		conditions = append(conditions, "r.date <= ?")
		args = append(args, q.EndDate)
	}
	if len(q.CategoryIDs) > 0 {
		placeholders := make([]string, len(q.CategoryIDs))
		for i, id := range q.CategoryIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, "r.category_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if q.Kind == "income" {
		conditions = append(conditions, "r.category_id = 'income'")
	} else if q.Kind == "expense" {
		conditions = append(conditions, "r.category_id <> 'income'")
	}
	if q.Keyword != "" {
		conditions = append(conditions, "(instr(r.tag, ?) > 0 OR instr(r.note, ?) > 0)")
		args = append(args, q.Keyword, q.Keyword)
	}
	if q.Tag != "" {
		// tag 欄位可能是「尼,午餐」這類逗號分隔的多選值，比對其中任一個完全相符。
		conditions = append(conditions, "instr(',' || r.tag || ',', ',' || ? || ',') > 0")
		args = append(args, q.Tag)
	}
	return " FROM records r LEFT JOIN spendCategories c ON c.id = r.category_id WHERE " +
		strings.Join(conditions, " AND "), args
}

func (a *Analyzer) queryAnalysis(ctx context.Context, query analysisQuery) (analysisResult, error) {
	result := analysisResult{Query: query}
	from, args := query.filter()
	totals := &result.Summary
	if err := a.db.QueryRowContext(ctx, "SELECT "+analysisTotalColumns+from, args...).Scan(
		&totals.Count, &totals.TotalCents, &totals.IncomeCents, &totals.ExpenseCents,
		&totals.MinCents, &totals.MaxCents, &totals.FirstDate, &totals.LastDate,
	); err != nil {
		return result, err
	}
	totals.BalanceCents = totals.IncomeCents - totals.ExpenseCents
	if column := analysisGroupColumns[query.GroupBy]; column != "" {
		rows, err := a.db.QueryContext(ctx,
			"SELECT "+column+", "+analysisTotalColumns+from+" GROUP BY "+column+
				" ORDER BY SUM(r.amount) DESC, "+column+" ASC LIMIT 101", args...)
		if err != nil {
			return result, err
		}
		defer rows.Close()
		for rows.Next() {
			var group analysisGroup
			if err := rows.Scan(&group.Key, &group.Count, &group.TotalCents,
				&group.IncomeCents, &group.ExpenseCents, &group.MinCents, &group.MaxCents,
				&group.FirstDate, &group.LastDate); err != nil {
				return result, err
			}
			group.BalanceCents = group.IncomeCents - group.ExpenseCents
			result.Groups = append(result.Groups, group)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return result, err
		}
		if len(result.Groups) > 100 {
			result.Groups = result.Groups[:100]
			result.GroupsTruncated = true
		}
	}
	if query.Details {
		detailArgs := append(append([]interface{}{}, args...), query.Limit)
		rows, err := a.db.QueryContext(ctx,
			`SELECT r.id, r.category_id, COALESCE(c.name, r.category_id), r.tag, r.note,
			 CAST(ROUND(r.amount * 100) AS INTEGER), r.date, COALESCE(r.created_at, '')`+
				from+" ORDER BY "+analysisOrders[query.Order]+" LIMIT ?", detailArgs...)
		if err != nil {
			return result, err
		}
		defer rows.Close()
		for rows.Next() {
			var rec analysisRecord
			if err := rows.Scan(&rec.ID, &rec.CategoryID, &rec.CategoryName, &rec.Tag,
				&rec.Note, &rec.AmountCents, &rec.Date, &rec.CreatedAt); err != nil {
				return result, err
			}
			result.Records = append(result.Records, rec)
		}
		if err := rows.Err(); err != nil {
			return result, err
		}
		result.RecordsTruncated = totals.Count > int64(len(result.Records))
	}
	return result, nil
}
