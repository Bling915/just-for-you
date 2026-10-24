package bootstrap

import (
	"context"
	"testing"
	"time"

	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type stubCurrentCouple struct{}

func (stubCurrentCouple) Execute(context.Context, coupleapp.GetCoupleQuery) (coupleapp.CurrentCouple, error) {
	return coupleapp.CurrentCouple{
		Couple:  coupledomain.NewCouple(12, 7, 8, time.Now()),
		Partner: coupleapp.PartnerProfile{UserID: 8, DisplayName: "小南"},
	}, nil
}

func TestMomentCoupleReaderTranslatesCurrentCoupleAtCompositionRoot(t *testing.T) {
	reader := newMomentCoupleReader(stubCurrentCouple{})

	result, err := reader.GetActiveCouple(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 12 || result.PartnerID != 8 {
		t.Fatalf("result=%+v", result)
	}
}
