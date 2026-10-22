package together

import (
	"context"
	"errors"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"time"
)

type DailyQuestionService struct {
	relationships RelationshipReader
	repo          DailyQuestionRepository
	generator     QuestionGenerator
	clock         Clock
}

func NewDailyQuestionService(r RelationshipReader, repo DailyQuestionRepository, g QuestionGenerator, c Clock) *DailyQuestionService {
	return &DailyQuestionService{r, repo, g, c}
}
func shanghaiDay(value time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	v := value.In(loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, loc)
}
func (s *DailyQuestionService) GetToday(ctx context.Context, userID int64) (domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	day := shanghaiDay(s.clock.Now())
	value, err := s.repo.FindDailyQuestion(ctx, rel.ID, day)
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.DailyQuestion{}, err
	}
	generated, genErr := s.generator.Generate(ctx, QuestionContext{CoupleID: rel.ID, Date: day})
	if genErr != nil {
		return domain.DailyQuestion{}, genErr
	}
	value, err = domain.NewDailyQuestion(rel.ID, day, generated.Text, generated.Source, s.clock.Now())
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	return s.repo.CreateDailyQuestion(ctx, value)
}

type SubmitDailyAnswerCommand struct {
	UserID, QuestionID int64
	Text               string
}

func (s *DailyQuestionService) SubmitAnswer(ctx context.Context, c SubmitDailyAnswerCommand) (domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveRelationship(ctx, c.UserID)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	questionID := c.QuestionID
	if questionID == 0 {
		q, e := s.GetToday(ctx, c.UserID)
		if e != nil {
			return domain.DailyQuestion{}, e
		}
		questionID = q.ID()
	}
	partner := rel.UserAID
	if partner == c.UserID {
		partner = rel.UserBID
	}
	return s.repo.SubmitDailyAnswer(ctx, questionID, c.UserID, partner, c.Text, s.clock.Now())
}
func (s *DailyQuestionService) List(ctx context.Context, userID int64) ([]domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveRelationship(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListDailyQuestions(ctx, rel.ID, 50)
}
