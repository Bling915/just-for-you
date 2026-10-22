package togethermysql

import (
	"context"
	"database/sql"
	moment "github.com/nanfang/just-for-you/server/internal/application/moment"
)

type HighlightReader struct{ db *sql.DB }

func NewHighlightReader(db *sql.DB) HighlightReader { return HighlightReader{db} }
func (r HighlightReader) ListPinned(ctx context.Context, coupleID int64, limit int) ([]moment.AnniversaryHighlight, error) {
	// 调用方可能传 displayLimit+1 来探测是否还有更多，所以这里要允许到 6。
	// 原先硬编码 limit>3 就回退成 3，导致调用方永远探测不到溢出。
	if limit < 1 || limit > 6 {
		limit = 6
	}
	rows, err := r.db.QueryContext(ctx, "SELECT id,title,anniversary_date,count_mode,repeats_yearly,pinned,cover_ref,message,system_kind FROM anniversaries WHERE couple_id=? AND pinned=TRUE ORDER BY CASE WHEN count_mode='countdown' THEN 0 ELSE 1 END,anniversary_date DESC,id LIMIT ?", coupleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []moment.AnniversaryHighlight{}
	for rows.Next() {
		var v moment.AnniversaryHighlight
		var date anniversaryDate
		if err = rows.Scan(&v.ID, &v.Title, &date, &v.CountMode, &v.RepeatsYearly, &v.Pinned, &v.CoverRef, &v.Message, &v.SystemKind); err != nil {
			return nil, err
		}
		v.Date = date.Time
		out = append(out, v)
	}
	return out, rows.Err()
}
