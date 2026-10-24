package bootstrap

import (
	"context"

	anniversaryapp "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	momentapp "github.com/nanfang/just-for-you/server/internal/moment/application"
)

type momentAnniversaryReader struct {
	reader anniversaryapp.PinnedAnniversaryReader
}

func newMomentAnniversaryReader(reader anniversaryapp.PinnedAnniversaryReader) momentAnniversaryReader {
	return momentAnniversaryReader{reader: reader}
}

func (reader momentAnniversaryReader) ListPinned(
	ctx context.Context,
	coupleID int64,
	limit int,
) ([]momentapp.AnniversaryHighlight, error) {
	items, err := reader.reader.ListPinned(ctx, coupleID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]momentapp.AnniversaryHighlight, 0, len(items))
	for _, item := range items {
		out = append(out, momentapp.AnniversaryHighlight{
			ID:            item.ID,
			Title:         item.Title,
			Date:          item.Date,
			CountMode:     item.CountMode,
			RepeatsYearly: item.RepeatsYearly,
			Pinned:        item.Pinned,
			CoverRef:      item.CoverRef,
			Message:       item.Message,
			SystemKind:    item.SystemKind,
		})
	}
	return out, nil
}
