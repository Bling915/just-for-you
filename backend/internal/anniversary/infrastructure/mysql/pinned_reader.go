package anniversarymysql

import (
	"context"
	"database/sql"

	app "github.com/nanfang/just-for-you/server/internal/anniversary/application"
)

type PinnedReader struct {
	db *sql.DB
}

func NewPinnedReader(db *sql.DB) PinnedReader {
	return PinnedReader{db: db}
}

func (reader PinnedReader) ListPinned(
	ctx context.Context,
	coupleID int64,
	limit int,
) ([]app.PinnedAnniversary, error) {
	if limit < 1 || limit > 6 {
		limit = 6
	}
	rows, err := reader.db.QueryContext(
		ctx,
		"SELECT id,title,anniversary_date,count_mode,repeats_yearly,pinned,cover_ref,message,system_kind FROM anniversaries WHERE couple_id=? AND pinned=TRUE ORDER BY CASE WHEN count_mode='countdown' THEN 0 ELSE 1 END,anniversary_date DESC,id LIMIT ?",
		coupleID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []app.PinnedAnniversary{}
	for rows.Next() {
		var item app.PinnedAnniversary
		var date anniversaryDate
		if err = rows.Scan(
			&item.ID,
			&item.Title,
			&date,
			&item.CountMode,
			&item.RepeatsYearly,
			&item.Pinned,
			&item.CoverRef,
			&item.Message,
			&item.SystemKind,
		); err != nil {
			return nil, err
		}
		item.Date = date.Time
		items = append(items, item)
	}
	return items, rows.Err()
}
