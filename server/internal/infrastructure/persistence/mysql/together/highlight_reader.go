package togethermysql

import (
	"context"
	"database/sql"
	moment "github.com/nanfang/just-for-you/server/internal/application/moment"
)

type HighlightReader struct{ db *sql.DB }

func NewHighlightReader(db *sql.DB) HighlightReader { return HighlightReader{db} }
func (r HighlightReader) ListPinned(ctx context.Context, coupleID int64, limit int) ([]moment.AnniversaryHighlight, error) {
	if limit < 1 || limit > 3 {
		limit = 3
	}
	rows, err := r.db.QueryContext(ctx, "SELECT id,title,anniversary_date,count_mode,repeats_yearly,pinned,cover_ref,message,system_kind FROM anniversaries WHERE couple_id=? AND pinned=TRUE ORDER BY CASE WHEN count_mode='countdown' THEN 0 ELSE 1 END,anniversary_date DESC,id LIMIT ?", coupleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []moment.AnniversaryHighlight{}
	for rows.Next() {
		var v moment.AnniversaryHighlight
		if err = rows.Scan(&v.ID, &v.Title, &v.Date, &v.CountMode, &v.RepeatsYearly, &v.Pinned, &v.CoverRef, &v.Message, &v.SystemKind); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
