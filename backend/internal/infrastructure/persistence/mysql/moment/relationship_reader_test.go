package infrastructure

import (
	"context"
	"testing"
	"time"

	coupleapplication "github.com/nanfang/just-for-you/server/internal/couple/application"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type currentCoupleService struct{}

func (currentCoupleService) Execute(context.Context, coupleapplication.GetCoupleQuery) (coupleapplication.CurrentCouple, error) {
	return coupleapplication.CurrentCouple{
		Couple:  coupledomain.NewCouple(12, 7, 8, time.Now()),
		Partner: coupleapplication.PartnerProfile{UserID: 8, DisplayName: "小南"},
	}, nil
}

func TestActiveCoupleReaderUsesStableCoupleInsideCurrentView(t *testing.T) {
	reader := NewActiveCoupleReader(currentCoupleService{})

	result, err := reader.GetActiveCouple(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 12 || result.PartnerID != 8 {
		t.Fatalf("result=%+v", result)
	}
}
