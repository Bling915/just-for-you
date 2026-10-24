package dailyquestion

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

type ListDailyQuestionsQuery struct {
	UserID int64
}

type ListDailyQuestions struct {
	relationships ActiveCoupleReader
	store         QuestionStore
}

func NewListDailyQuestions(
	relationships ActiveCoupleReader,
	store QuestionStore,
) *ListDailyQuestions {
	return &ListDailyQuestions{relationships: relationships, store: store}
}

func (usecase *ListDailyQuestions) Execute(
	ctx context.Context,
	query ListDailyQuestionsQuery,
) ([]domain.DailyQuestion, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	return usecase.store.List(ctx, relationship.ID, 50)
}
