package moment

import (
	"context"
	"fmt"
	"time"
)

type ReadModels struct {
	relationships ActiveRelationshipReader
	repo          ReadRepository
}

func NewReadModels(r ActiveRelationshipReader, repo ReadRepository) *ReadModels {
	return &ReadModels{r, repo}
}
func monthRange(month string) (time.Time, time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	var y, m int
	if _, e := fmt.Sscanf(month, "%d-%d", &y, &m); e != nil || m < 1 || m > 12 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month")
	}
	start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 1, 0), nil
}
func (r *ReadModels) Calendar(ctx context.Context, userID int64, month string) ([]CalendarDay, error) {
	rel, e := r.relationships.GetActiveRelationship(ctx, userID)
	if e != nil {
		return nil, e
	}
	from, to, e := monthRange(month)
	if e != nil {
		return nil, e
	}
	return r.repo.Calendar(ctx, rel.ID, from, to)
}
func (r *ReadModels) Map(ctx context.Context, userID int64, from, to time.Time) ([]MapPoint, error) {
	rel, e := r.relationships.GetActiveRelationship(ctx, userID)
	if e != nil {
		return nil, e
	}
	return r.repo.Map(ctx, rel.ID, from, to, 500)
}
func (r *ReadModels) Report(ctx context.Context, userID int64, month string) (MonthlyReport, error) {
	rel, e := r.relationships.GetActiveRelationship(ctx, userID)
	if e != nil {
		return MonthlyReport{}, e
	}
	from, to, e := monthRange(month)
	if e != nil {
		return MonthlyReport{}, e
	}
	v, e := r.repo.MonthlyReport(ctx, rel.ID, from, to)
	v.Month = month
	return v, e
}
