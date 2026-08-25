package store

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"context"
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type CaseQuery struct {
	Keyword  string
	Owner    string
	Status   domain.CaseStatus
	DateFrom string
	DateTo   string
	Sort     string
	Order    string
	Cursor   string
	Limit    int
}

type CasePage struct {
	Items      []domain.AcceptanceCase
	Total      int
	NextCursor string
}

type caseCursor struct {
	Sort  string `json:"sort"`
	Order string `json:"order"`
	Value string `json:"value"`
	ID    string `json:"id"`
}

func (r *DiskRepository) SearchCases(ctx context.Context, query CaseQuery) (CasePage, error) {
	items, err := r.List(ctx)
	if err != nil {
		return CasePage{}, err
	}
	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	owner := strings.ToLower(strings.TrimSpace(query.Owner))
	filtered := items[:0]
	for _, item := range items {
		if keyword != "" && !strings.Contains(strings.ToLower(item.Title), keyword) && !strings.Contains(strings.ToLower(item.Organization), keyword) {
			continue
		}
		if owner != "" && strings.ToLower(strings.TrimSpace(item.Owner)) != owner {
			continue
		}
		if query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.DateFrom != "" && item.TargetPublishDate < query.DateFrom {
			continue
		}
		if query.DateTo != "" && item.TargetPublishDate > query.DateTo {
			continue
		}
		filtered = append(filtered, item)
	}
	total := len(filtered)
	sort.Slice(filtered, func(i, j int) bool { return caseBefore(filtered[i], filtered[j], query.Sort, query.Order) })
	start := 0
	if query.Cursor != "" {
		cursor, err := decodeCaseCursor(query.Cursor)
		if err != nil || cursor.Sort != query.Sort || cursor.Order != query.Order {
			return CasePage{}, domain.ValidationErrors{{Field: "cursor", Message: "分页游标无效或与当前排序不匹配"}}
		}
		found := false
		for index := range filtered {
			if cursor.Value == caseSortValue(filtered[index], query.Sort) && cursor.ID == filtered[index].ID {
				start = index + 1
				found = true
				break
			}
		}
		if !found {
			return CasePage{}, domain.ValidationErrors{{Field: "cursor", Message: "分页游标已失效，请重新查询"}}
		}
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	page := CasePage{Items: append([]domain.AcceptanceCase(nil), filtered[start:end]...), Total: total}
	if end < len(filtered) && end > 0 {
		last := filtered[end-1]
		page.NextCursor = encodeCaseCursor(caseCursor{query.Sort, query.Order, caseSortValue(last, query.Sort), last.ID})
	}
	return page, nil
}

func caseSortValue(item domain.AcceptanceCase, field string) string {
	if field == "target_publish_date" {
		return item.TargetPublishDate
	}
	return item.UpdatedAt.UTC().Format(time.RFC3339Nano)
}

func caseBefore(left, right domain.AcceptanceCase, field, order string) bool {
	lv, rv := caseSortValue(left, field), caseSortValue(right, field)
	if lv == rv {
		return left.ID < right.ID
	}
	if order == "asc" {
		return lv < rv
	}
	return lv > rv
}

func encodeCaseCursor(cursor caseCursor) string {
	b, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(b)
}
func decodeCaseCursor(value string) (caseCursor, error) {
	var cursor caseCursor
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		err = json.Unmarshal(b, &cursor)
	}
	return cursor, err
}
