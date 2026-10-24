package dailyquestion

import (
	"context"
	"errors"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

type GetDailyQuestionQuery struct {
	UserID int64
}

type GetDailyQuestion struct {
	relationships ActiveCoupleReader
	store         QuestionStore
	generator     QuestionGenerator
	clock         Clock
	contextReader QuestionContextReader
}

func NewGetDailyQuestion(
	relationships ActiveCoupleReader,
	store QuestionStore,
	generator QuestionGenerator,
	clock Clock,
	readers ...QuestionContextReader,
) *GetDailyQuestion {
	var contextReader QuestionContextReader
	if len(readers) > 0 {
		contextReader = readers[0]
	}
	return &GetDailyQuestion{
		relationships: relationships,
		store:         store,
		generator:     generator,
		clock:         clock,
		contextReader: contextReader,
	}
}

func (usecase *GetDailyQuestion) Execute(
	ctx context.Context,
	query GetDailyQuestionQuery,
) (domain.DailyQuestion, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, query.UserID)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	day := shanghaiDay(usecase.clock.Now())
	question, err := usecase.store.Find(ctx, relationship.ID, day)
	if err == nil {
		return question, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.DailyQuestion{}, err
	}
	questionContext := QuestionContext{CoupleID: relationship.ID, Date: day}
	if usecase.contextReader != nil {
		if values, contextErr := usecase.contextReader.RecentContextMoments(
			ctx,
			relationship.ID,
			20,
		); contextErr == nil {
			questionContext = (QuestionContextBuilder{}).Build(
				relationship.ID,
				day,
				values,
			)
		}
	}
	generated, err := usecase.generator.Generate(ctx, questionContext)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	question, err = domain.NewDailyQuestion(
		relationship.ID,
		day,
		generated.Text,
		generated.Source,
		usecase.clock.Now(),
	)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	question = question.WithContextVersion(generated.ContextVersion)
	return usecase.store.Create(ctx, question)
}

func shanghaiDay(value time.Time) time.Time {
	location, _ := time.LoadLocation("Asia/Shanghai")
	local := value.In(location)
	return time.Date(
		local.Year(),
		local.Month(),
		local.Day(),
		0,
		0,
		0,
		0,
		location,
	)
}
