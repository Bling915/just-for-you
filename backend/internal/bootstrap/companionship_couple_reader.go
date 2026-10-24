package bootstrap

import (
	"context"

	companionshipapp "github.com/nanfang/just-for-you/server/internal/companionship/application"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
)

type companionshipCoupleReader struct {
	service currentCoupleService
}

func newCompanionshipCoupleReader(service currentCoupleService) companionshipCoupleReader {
	return companionshipCoupleReader{service: service}
}

func (reader companionshipCoupleReader) GetActiveCouple(
	ctx context.Context,
	userID int64,
) (companionshipapp.ActiveCouple, error) {
	current, err := reader.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return companionshipapp.ActiveCouple{}, err
	}
	relationship := current.Couple
	return companionshipapp.ActiveCouple{
		ID:      relationship.ID(),
		UserAID: relationship.UserAID(),
		UserBID: relationship.UserBID(),
	}, nil
}
