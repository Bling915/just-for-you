package bootstrap

import (
	"context"

	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	momentapp "github.com/nanfang/just-for-you/server/internal/moment/application"
)

type currentCoupleService interface {
	Execute(context.Context, coupleapp.GetCoupleQuery) (coupleapp.CurrentCouple, error)
}

type momentCoupleReader struct {
	service currentCoupleService
}

func newMomentCoupleReader(service currentCoupleService) momentCoupleReader {
	return momentCoupleReader{service: service}
}

func (r momentCoupleReader) GetActiveCouple(ctx context.Context, userID int64) (momentapp.ActiveCouple, error) {
	current, err := r.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return momentapp.ActiveCouple{}, err
	}
	couple := current.Couple
	partnerID := couple.UserAID()
	if partnerID == userID {
		partnerID = couple.UserBID()
	}
	return momentapp.ActiveCouple{ID: couple.ID(), PartnerID: partnerID}, nil
}
