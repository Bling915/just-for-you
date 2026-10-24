package together

import (
	"context"
	"errors"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"time"
)

type DailyQuestionService struct {
	relationships ActiveCoupleReader
	repo          DailyQuestionStore
	generator     QuestionGenerator
	clock         Clock
	contextReader QuestionContextReader
}

func NewDailyQuestionService(r ActiveCoupleReader, repo DailyQuestionStore, g QuestionGenerator, c Clock, readers ...QuestionContextReader) *DailyQuestionService {
	var reader QuestionContextReader
	if len(readers) > 0 {
		reader = readers[0]
	}
	return &DailyQuestionService{relationships: r, repo: repo, generator: g, clock: c, contextReader: reader}
}
func shanghaiDay(value time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	v := value.In(loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, loc)
}
func (s *DailyQuestionService) GetToday(ctx context.Context, userID int64) (domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveCouple(ctx, userID)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	day := shanghaiDay(s.clock.Now())
	value, err := s.repo.Find(ctx, rel.ID, day)
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.DailyQuestion{}, err
	}
	questionContext := QuestionContext{CoupleID: rel.ID, Date: day}
	if s.contextReader != nil {
		if values, e := s.contextReader.RecentContextMoments(ctx, rel.ID, 20); e == nil {
			questionContext = (QuestionContextBuilder{}).Build(rel.ID, day, values)
		}
	}
	generated, genErr := s.generator.Generate(ctx, questionContext)
	if genErr != nil {
		return domain.DailyQuestion{}, genErr
	}
	value, err = domain.NewDailyQuestion(rel.ID, day, generated.Text, generated.Source, s.clock.Now())
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	value = value.WithContextVersion(generated.ContextVersion)
	return s.repo.Create(ctx, value)
}

type SubmitDailyAnswerCommand struct {
	UserID, QuestionID int64
	Text               string
}

func (s *DailyQuestionService) SubmitAnswer(ctx context.Context, c SubmitDailyAnswerCommand) (domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveCouple(ctx, c.UserID)
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
	return s.repo.SubmitAnswer(ctx, questionID, c.UserID, partner, c.Text, s.clock.Now())
}
func (s *DailyQuestionService) List(ctx context.Context, userID int64) ([]domain.DailyQuestion, error) {
	rel, err := s.relationships.GetActiveCouple(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, rel.ID, 50)
}
