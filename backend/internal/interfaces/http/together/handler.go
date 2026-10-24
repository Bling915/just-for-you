package togetherhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	shared "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
)

type GetDailyQuestionExecutor interface {
	GetToday(context.Context, int64) (domain.DailyQuestion, error)
}

type SubmitDailyAnswerExecutor interface {
	SubmitAnswer(context.Context, app.SubmitDailyAnswerCommand) (domain.DailyQuestion, error)
}

type ListDailyQuestionsExecutor interface {
	List(context.Context, int64) ([]domain.DailyQuestion, error)
}

type Dependencies struct {
	GetToday     GetDailyQuestionExecutor
	SubmitAnswer SubmitDailyAnswerExecutor
	List         ListDailyQuestionsExecutor
}

type Handler struct {
	dependencies Dependencies
}

func NewHandler(dependencies Dependencies) *Handler {
	return &Handler{dependencies: dependencies}
}

func (*Handler) RegisterPublic(*gin.Engine) {}

func (handler *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/daily-question", handler.getToday)
	group.PUT("/daily-question/answer", handler.submitAnswer)
	group.GET("/daily-questions", handler.list)
}

func (handler *Handler) getToday(c *gin.Context) {
	item, err := handler.dependencies.GetToday.GetToday(c, shared.UserID(c))
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusOK, dailyQuestionDTO(item, shared.UserID(c)))
}

func (handler *Handler) submitAnswer(c *gin.Context) {
	var input struct {
		QuestionID int64  `json:"questionId"`
		Text       string `json:"text"`
	}
	if c.ShouldBindJSON(&input) != nil {
		shared.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "回答内容不正确")
		return
	}
	item, err := handler.dependencies.SubmitAnswer.SubmitAnswer(
		c,
		app.SubmitDailyAnswerCommand{
			UserID:     shared.UserID(c),
			QuestionID: input.QuestionID,
			Text:       input.Text,
		},
	)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusOK, dailyQuestionDTO(item, shared.UserID(c)))
}

func (handler *Handler) list(c *gin.Context) {
	items, err := handler.dependencies.List.List(c, shared.UserID(c))
	if err != nil {
		fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, dailyQuestionDTO(item, shared.UserID(c)))
	}
	shared.WriteData(c, http.StatusOK, out)
}

func dailyQuestionDTO(value domain.DailyQuestion, userID int64) gin.H {
	answers := value.Answers()
	var mine any
	if answer, ok := answers[userID]; ok {
		mine = gin.H{"text": answer.Text, "submittedAt": answer.SubmittedAt}
	}
	var partner any
	if value.Unlocked() {
		for id, answer := range answers {
			if id != userID {
				partner = gin.H{"text": answer.Text, "submittedAt": answer.SubmittedAt}
			}
		}
	}
	return gin.H{
		"id":            value.ID(),
		"date":          value.Date(),
		"question":      value.Text(),
		"source":        value.Source(),
		"unlocked":      value.Unlocked(),
		"myAnswer":      mine,
		"partnerAnswer": partner,
		"answered":      mine != nil,
	}
}

func fail(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(err, domain.ErrInvalidDailyAnswer):
		status, code, message = http.StatusBadRequest, "INVALID_DAILY_ANSWER", "回答不能为空且不能超过 500 个字"
	case errors.Is(err, domain.ErrDailyQuestionFrozen):
		status, code, message = http.StatusConflict, "DAILY_QUESTION_FROZEN", "双方已回答，今天的答案已经珍藏"
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "内容不存在"
	case errors.Is(err, coupledomain.ErrNoActiveCouple):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定关系"
	default:
		_ = c.Error(err)
	}
	shared.WriteError(c, status, code, message)
}
