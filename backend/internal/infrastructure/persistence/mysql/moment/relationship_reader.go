package infrastructure

import (
	"context"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	coupleapplication "github.com/nanfang/just-for-you/server/internal/couple/application"
)

type CoupleService interface {
	Execute(context.Context, coupleapplication.GetCoupleQuery) (coupleapplication.CurrentCouple, error)
}
type ActiveCoupleReader struct{ service CoupleService }

func NewActiveCoupleReader(service CoupleService) ActiveCoupleReader {
	return ActiveCoupleReader{service: service}
}
func (r ActiveCoupleReader) GetActiveCouple(ctx context.Context, userID int64) (application.ActiveCouple, error) {
	current, err := r.service.Execute(ctx, coupleapplication.GetCoupleQuery{UserID: userID})
	if err != nil {
		return application.ActiveCouple{}, err
	}
	relationship := current.Couple
	partner := relationship.UserAID()
	if partner == userID {
		partner = relationship.UserBID()
	}
	return application.ActiveCouple{ID: relationship.ID(), PartnerID: partner}, nil
}
