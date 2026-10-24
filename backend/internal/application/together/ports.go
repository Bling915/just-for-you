package together

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type ActiveCouple struct {
	ID      int64
	UserAID int64
	UserBID int64
}

type ActiveCoupleReader interface {
	GetActiveCouple(context.Context, int64) (ActiveCouple, error)
}

type Clock interface {
	Now() time.Time
}

type GeneratedQuestion struct {
	Text           string
	Source         string
	ContextVersion string
}

type QuestionContext struct {
	CoupleID int64
	Date     time.Time
	Themes   []string
}

type QuestionGenerator interface {
	Generate(context.Context, QuestionContext) (GeneratedQuestion, error)
}

type AIPersonalizationReader interface {
	AIPersonalizationEnabled(context.Context, int64) (bool, error)
}

type QuestionContextReader interface {
	RecentContextMoments(context.Context, int64, int) ([]ContextMoment, error)
}

type DailyQuestionStore interface {
	Find(context.Context, int64, time.Time) (domain.DailyQuestion, error)
	FindByID(context.Context, int64, int64) (domain.DailyQuestion, error)
	Create(context.Context, domain.DailyQuestion) (domain.DailyQuestion, error)
	SubmitAnswer(context.Context, int64, int64, int64, string, time.Time) (domain.DailyQuestion, error)
	List(context.Context, int64, int) ([]domain.DailyQuestion, error)
}
