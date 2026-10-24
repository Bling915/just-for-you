package moment

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"testing"
	"time"
)

func TestResurfacedMemoryPrioritizesLastYearToday(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	lastMonth := domain.RestoreMomentDetails(domain.RestoreInput{ID: 2, AuthorID: 1, CoupleID: 9, Text: "上个月", SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, Kind: domain.KindStandard, OccurredAt: now.AddDate(0, -1, 0), CreatedAt: now.AddDate(0, -1, 0), UpdatedAt: now.AddDate(0, -1, 0)})
	lastYear := domain.RestoreMomentDetails(domain.RestoreInput{ID: 1, AuthorID: 2, CoupleID: 9, Text: "去年今天", SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, Kind: domain.KindStandard, OccurredAt: now.AddDate(-1, 0, 0), CreatedAt: now.AddDate(-1, 0, 0), UpdatedAt: now.AddDate(-1, 0, 0)})
	repo := &memoryMoments{resurfaceCandidates: []domain.Moment{lastMonth, lastYear}}
	useCase := NewGetResurfacedMemoryUseCase(fakeCouples{active: ActiveCouple{ID: 9, PartnerID: 2}}, repo, fakeStorage{}, fixedClock{now})
	result, err := useCase.Execute(context.Background(), GetResurfacedMemoryQuery{UserID: 1})
	if err != nil || result.Memory == nil || result.Memory.ID() != 1 || result.Reason != "去年今天" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
func TestResurfacedMemoryReturnsEmptyWithoutCandidate(t *testing.T) {
	useCase := NewGetResurfacedMemoryUseCase(fakeCouples{active: ActiveCouple{ID: 9, PartnerID: 2}}, &memoryMoments{}, fakeStorage{}, fixedClock{time.Now()})
	result, err := useCase.Execute(context.Background(), GetResurfacedMemoryQuery{UserID: 1})
	if err != nil || result.Memory != nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
