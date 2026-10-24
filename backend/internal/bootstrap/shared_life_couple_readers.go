package bootstrap

import (
	"context"

	anniversaryapp "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	togetherapp "github.com/nanfang/just-for-you/server/internal/application/together"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	todoapp "github.com/nanfang/just-for-you/server/internal/todo/application"
	wishapp "github.com/nanfang/just-for-you/server/internal/wish/application"
)

type anniversaryCoupleReader struct {
	service currentCoupleService
}

func newAnniversaryCoupleReader(service currentCoupleService) anniversaryCoupleReader {
	return anniversaryCoupleReader{service: service}
}

func (reader anniversaryCoupleReader) GetActiveCouple(
	ctx context.Context,
	userID int64,
) (anniversaryapp.ActiveCouple, error) {
	current, err := reader.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return anniversaryapp.ActiveCouple{}, err
	}
	couple := current.Couple
	return anniversaryapp.ActiveCouple{
		ID:      couple.ID(),
		UserAID: couple.UserAID(),
		UserBID: couple.UserBID(),
		BoundAt: couple.BoundAt(),
	}, nil
}

type todoCoupleReader struct {
	service currentCoupleService
}

func newTodoCoupleReader(service currentCoupleService) todoCoupleReader {
	return todoCoupleReader{service: service}
}

func (reader todoCoupleReader) GetActiveCouple(
	ctx context.Context,
	userID int64,
) (todoapp.ActiveCouple, error) {
	current, err := reader.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return todoapp.ActiveCouple{}, err
	}
	return todoapp.ActiveCouple{ID: current.Couple.ID()}, nil
}

type wishCoupleReader struct {
	service currentCoupleService
}

func newWishCoupleReader(service currentCoupleService) wishCoupleReader {
	return wishCoupleReader{service: service}
}

func (reader wishCoupleReader) GetActiveCouple(
	ctx context.Context,
	userID int64,
) (wishapp.ActiveCouple, error) {
	current, err := reader.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return wishapp.ActiveCouple{}, err
	}
	return wishapp.ActiveCouple{ID: current.Couple.ID()}, nil
}

type dailyQuestionCoupleReader struct {
	service currentCoupleService
}

func newDailyQuestionCoupleReader(service currentCoupleService) dailyQuestionCoupleReader {
	return dailyQuestionCoupleReader{service: service}
}

func (reader dailyQuestionCoupleReader) GetActiveCouple(
	ctx context.Context,
	userID int64,
) (togetherapp.ActiveCouple, error) {
	current, err := reader.service.Execute(ctx, coupleapp.GetCoupleQuery{UserID: userID})
	if err != nil {
		return togetherapp.ActiveCouple{}, err
	}
	couple := current.Couple
	return togetherapp.ActiveCouple{
		ID:      couple.ID(),
		UserAID: couple.UserAID(),
		UserBID: couple.UserBID(),
	}, nil
}
