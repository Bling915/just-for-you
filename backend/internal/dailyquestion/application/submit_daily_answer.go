package dailyquestion

import (
	"context"

	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

type SubmitDailyAnswerCommand struct {
	UserID     int64
	QuestionID int64
	Text       string
}

type SubmitDailyAnswer struct {
	relationships ActiveCoupleReader
	answers       AnswerSubmissionStore
	clock         Clock
	today         *GetDailyQuestion
}

func NewSubmitDailyAnswer(
	relationships ActiveCoupleReader,
	answers AnswerSubmissionStore,
	clock Clock,
	today *GetDailyQuestion,
) *SubmitDailyAnswer {
	return &SubmitDailyAnswer{
		relationships: relationships,
		answers:       answers,
		clock:         clock,
		today:         today,
	}
}

func (usecase *SubmitDailyAnswer) Execute(
	ctx context.Context,
	command SubmitDailyAnswerCommand,
) (domain.DailyQuestion, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	questionID := command.QuestionID
	if questionID == 0 {
		question, err := usecase.today.Execute(
			ctx,
			GetDailyQuestionQuery{UserID: command.UserID},
		)
		if err != nil {
			return domain.DailyQuestion{}, err
		}
		questionID = question.ID()
	}
	question, err := usecase.answers.Submit(
		ctx,
		SubmitAnswerTransaction{
			QuestionID: questionID,
			UserID:     command.UserID,
			PartnerID:  relationship.PartnerOf(command.UserID),
			Text:       command.Text,
			Now:        usecase.clock.Now(),
		},
	)
	return question, err
}
