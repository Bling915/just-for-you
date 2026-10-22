package togetherhttp

import (
	"errors"
	"github.com/gin-gonic/gin"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	relationship "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	shared "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	s     *app.Service
	daily *app.DailyQuestionService
}

func NewHandler(s *app.Service, daily ...*app.DailyQuestionService) *Handler {
	h := &Handler{s: s}
	if len(daily) > 0 {
		h.daily = daily[0]
	}
	return h
}
func (*Handler) RegisterPublic(*gin.Engine) {}
func (h *Handler) RegisterProtected(g *gin.RouterGroup) {
	g.GET("/anniversaries", h.listAnniversaries)
	g.POST("/anniversaries", h.createAnniversary)
	g.GET("/anniversaries/:id", h.getAnniversary)
	g.PATCH("/anniversaries/:id", h.updateAnniversary)
	g.POST("/anniversaries/:id/cover-upload", h.requestCoverUpload)
	g.DELETE("/anniversaries/:id", h.deleteAnniversary)
	g.GET("/todos", h.listTodos)
	g.POST("/todos", h.createTodo)
	g.PATCH("/todos/:id", h.updateTodo)
	g.DELETE("/todos/:id", h.deleteTodo)
	g.GET("/wishes", h.listWishes)
	g.POST("/wishes", h.createWish)
	g.PATCH("/wishes/:id", h.updateWish)
	g.DELETE("/wishes/:id", h.deleteWish)
	g.GET("/daily-question", h.getDailyQuestion)
	g.PUT("/daily-question/answer", h.submitDailyAnswer)
	g.GET("/daily-questions", h.listDailyQuestions)
}
func (h *Handler) getDailyQuestion(c *gin.Context) {
	if h.daily == nil {
		shared.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	value, e := h.daily.GetToday(c, shared.UserID(c))
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 200, dailyQuestionDTO(value, shared.UserID(c)))
}
func (h *Handler) submitDailyAnswer(c *gin.Context) {
	if h.daily == nil {
		shared.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	var input struct {
		QuestionID int64  `json:"questionId"`
		Text       string `json:"text"`
	}
	if c.ShouldBindJSON(&input) != nil {
		shared.WriteError(c, 400, "INVALID_ARGUMENT", "回答内容不正确")
		return
	}
	value, e := h.daily.SubmitAnswer(c, app.SubmitDailyAnswerCommand{UserID: shared.UserID(c), QuestionID: input.QuestionID, Text: input.Text})
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 200, dailyQuestionDTO(value, shared.UserID(c)))
}
func (h *Handler) listDailyQuestions(c *gin.Context) {
	if h.daily == nil {
		shared.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	values, e := h.daily.List(c, shared.UserID(c))
	if e != nil {
		fail(c, e)
		return
	}
	out := make([]gin.H, 0, len(values))
	for _, value := range values {
		out = append(out, dailyQuestionDTO(value, shared.UserID(c)))
	}
	shared.WriteData(c, 200, out)
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
	return gin.H{"id": value.ID(), "date": value.Date(), "question": value.Text(), "source": value.Source(), "unlocked": value.Unlocked(), "myAnswer": mine, "partnerAnswer": partner, "answered": mine != nil}
}
func id(c *gin.Context) (int64, error) { return strconv.ParseInt(c.Param("id"), 10, 64) }
func (h *Handler) listAnniversaries(c *gin.Context) {
	v, e := h.s.ListAnniversaries(c, shared.UserID(c))
	if e != nil {
		fail(c, e)
		return
	}
	out := make([]gin.H, 0, len(v))
	for _, item := range v {
		dto, err := h.anniversaryDTO(c, item)
		if err != nil {
			fail(c, err)
			return
		}
		out = append(out, dto)
	}
	shared.WriteData(c, 200, out)
}
func (h *Handler) createAnniversary(c *gin.Context) {
	var x anniversaryInput
	if c.ShouldBindJSON(&x) != nil {
		shared.WriteError(c, 400, "INVALID_ARGUMENT", "纪念日内容不正确")
		return
	}
	v, e := h.s.CreateAnniversary(c, shared.UserID(c), x.draft())
	if e != nil {
		fail(c, e)
		return
	}
	dto, e := h.anniversaryDTO(c, v)
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 201, dto)
}

type anniversaryInput struct {
	Title     string           `json:"title"`
	Date      time.Time        `json:"date"`
	CountMode domain.CountMode `json:"countMode"`
	Repeats   bool             `json:"repeatsYearly"`
	Pinned    bool             `json:"pinned"`
	CoverRef  string           `json:"coverRef"`
	Message   string           `json:"message"`
}

func (x anniversaryInput) draft() domain.AnniversaryDraft {
	if x.CountMode == "" {
		x.CountMode = domain.CountModeCountdown
	}
	return domain.AnniversaryDraft{Title: x.Title, Date: x.Date, CountMode: x.CountMode, RepeatsYearly: x.Repeats, Pinned: x.Pinned, CoverRef: x.CoverRef, Message: x.Message}
}
func (h *Handler) getAnniversary(c *gin.Context) {
	v, e := id(c)
	if e == nil {
		var out domain.Anniversary
		out, e = h.s.GetAnniversary(c, shared.UserID(c), v)
		if e == nil {
			dto, err := h.anniversaryDTO(c, out)
			if err != nil {
				fail(c, err)
				return
			}
			shared.WriteData(c, 200, dto)
			return
		}
	}
	fail(c, e)
}
func (h *Handler) updateAnniversary(c *gin.Context) {
	var x anniversaryInput
	v, e := id(c)
	if c.ShouldBindJSON(&x) != nil {
		e = domain.ErrInvalidTitle
	}
	if e == nil {
		var out domain.Anniversary
		out, e = h.s.UpdateAnniversary(c, shared.UserID(c), v, x.draft())
		if e == nil {
			dto, err := h.anniversaryDTO(c, out)
			if err != nil {
				fail(c, err)
				return
			}
			shared.WriteData(c, 200, dto)
			return
		}
	}
	fail(c, e)
}
func (h *Handler) requestCoverUpload(c *gin.Context) {
	v, e := id(c)
	var x struct {
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
	}
	if c.ShouldBindJSON(&x) != nil {
		e = domain.ErrInvalidCover
	}
	if e == nil {
		var g app.CoverUploadGrant
		g, e = h.s.RequestCoverUpload(c, shared.UserID(c), v, x.ContentType, x.Size)
		if e == nil {
			shared.WriteData(c, 201, g)
			return
		}
	}
	fail(c, e)
}
func (h *Handler) deleteAnniversary(c *gin.Context) {
	v, e := id(c)
	if e == nil {
		e = h.s.DeleteAnniversary(c, shared.UserID(c), v)
	}
	finish(c, e)
}
func (h *Handler) listTodos(c *gin.Context) {
	v, e := h.s.ListTodos(c, shared.UserID(c))
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 200, todosDTO(v))
}
func (h *Handler) createTodo(c *gin.Context) {
	var x struct {
		Title string     `json:"title"`
		DueAt *time.Time `json:"dueAt"`
	}
	if c.ShouldBindJSON(&x) != nil {
		shared.WriteError(c, 400, "INVALID_ARGUMENT", "待办内容不正确")
		return
	}
	v, e := h.s.CreateTodo(c, shared.UserID(c), x.Title, x.DueAt)
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 201, todoDTO(v))
}
func (h *Handler) updateTodo(c *gin.Context) {
	var x struct {
		Completed bool `json:"completed"`
	}
	v, e := id(c)
	if c.ShouldBindJSON(&x) != nil {
		e = domain.ErrInvalidTitle
	}
	if e == nil {
		var out domain.Todo
		out, e = h.s.SetTodoCompleted(c, shared.UserID(c), v, x.Completed)
		if e == nil {
			shared.WriteData(c, 200, todoDTO(out))
			return
		}
	}
	fail(c, e)
}
func (h *Handler) deleteTodo(c *gin.Context) {
	v, e := id(c)
	if e == nil {
		e = h.s.DeleteTodo(c, shared.UserID(c), v)
	}
	finish(c, e)
}
func (h *Handler) listWishes(c *gin.Context) {
	v, e := h.s.ListWishes(c, shared.UserID(c))
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 200, wishesDTO(v))
}
func (h *Handler) createWish(c *gin.Context) {
	var x struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&x) != nil {
		shared.WriteError(c, 400, "INVALID_ARGUMENT", "愿望内容不正确")
		return
	}
	v, e := h.s.CreateWish(c, shared.UserID(c), x.Title, x.Description)
	if e != nil {
		fail(c, e)
		return
	}
	shared.WriteData(c, 201, wishDTO(v))
}
func (h *Handler) updateWish(c *gin.Context) {
	var x struct {
		Completed      bool   `json:"completed"`
		LinkedMomentID *int64 `json:"linkedMomentId"`
	}
	v, e := id(c)
	if c.ShouldBindJSON(&x) != nil {
		e = domain.ErrInvalidTitle
	}
	if e == nil {
		var out domain.Wish
		out, e = h.s.SetWishCompleted(c, shared.UserID(c), v, x.Completed, x.LinkedMomentID)
		if e == nil {
			shared.WriteData(c, 200, wishDTO(out))
			return
		}
	}
	fail(c, e)
}
func (h *Handler) deleteWish(c *gin.Context) {
	v, e := id(c)
	if e == nil {
		e = h.s.DeleteWish(c, shared.UserID(c), v)
	}
	finish(c, e)
}
func finish(c *gin.Context, e error) {
	if e != nil {
		fail(c, e)
		return
	}
	c.Status(http.StatusNoContent)
}
func fail(c *gin.Context, e error) {
	status, code, msg := 500, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(e, domain.ErrInvalidTitle), errors.Is(e, domain.ErrInvalidDate), errors.Is(e, domain.ErrInvalidDescription), errors.Is(e, domain.ErrInvalidMessage), errors.Is(e, domain.ErrInvalidCountMode):
		status, code, msg = 400, "INVALID_ARGUMENT", "内容不正确"
	case errors.Is(e, domain.ErrInvalidCover):
		status, code, msg = 400, "INVALID_COVER", "封面无效或已过期"
	case errors.Is(e, domain.ErrInvalidMoment):
		status, code, msg = 400, "INVALID_MOMENT", "关联动态不正确"
	case errors.Is(e, domain.ErrNotCreator):
		status, code, msg = 403, "NOT_CREATOR", "只能删除自己创建的内容"
	case errors.Is(e, domain.ErrSystemAnniversary):
		status, code, msg = 403, "SYSTEM_ANNIVERSARY", "系统纪念日不能删除"
	case errors.Is(e, domain.ErrInvalidDailyAnswer):
		status, code, msg = 400, "INVALID_DAILY_ANSWER", "回答不能为空且不能超过 500 个字"
	case errors.Is(e, domain.ErrDailyQuestionFrozen):
		status, code, msg = 409, "DAILY_QUESTION_FROZEN", "双方已回答，今天的答案已经珍藏"
	case errors.Is(e, domain.ErrNotFound):
		status, code, msg = 404, "NOT_FOUND", "内容不存在"
	case errors.Is(e, relationship.ErrNoActiveRelationship):
		status, code, msg = 404, "NO_ACTIVE_COUPLE", "还没有绑定关系"
	default:
		_ = c.Error(e)
	}
	shared.WriteError(c, status, code, msg)
}

func (h *Handler) anniversaryDTO(c *gin.Context, v domain.Anniversary) (gin.H, error) {
	url, err := h.s.CoverURL(c, v.CoverRef())
	if err != nil {
		return nil, err
	}
	return gin.H{"id": v.ID(), "creatorUserId": v.CreatorID(), "title": v.Title(), "date": v.Date(), "countMode": v.CountMode(), "repeatsYearly": v.RepeatsYearly(), "pinned": v.Pinned(), "coverRef": v.CoverRef(), "coverUrl": url, "message": v.Message(), "systemKind": v.SystemKind(), "createdAt": v.CreatedAt(), "updatedAt": v.UpdatedAt()}, nil
}
func todoDTO(v domain.Todo) gin.H {
	return gin.H{"id": v.ID(), "creatorUserId": v.CreatorID(), "completedByUserId": v.CompletedBy(), "title": v.Title(), "dueAt": v.DueAt(), "status": v.Status(), "completedAt": v.CompletedAt(), "createdAt": v.CreatedAt(), "updatedAt": v.UpdatedAt()}
}
func todosDTO(values []domain.Todo) []gin.H {
	out := make([]gin.H, 0, len(values))
	for _, v := range values {
		out = append(out, todoDTO(v))
	}
	return out
}
func wishDTO(v domain.Wish) gin.H {
	return gin.H{"id": v.ID(), "creatorUserId": v.CreatorID(), "completedByUserId": v.CompletedBy(), "linkedMomentId": v.LinkedMomentID(), "title": v.Title(), "description": v.Description(), "status": v.Status(), "completedAt": v.CompletedAt(), "createdAt": v.CreatedAt(), "updatedAt": v.UpdatedAt()}
}
func wishesDTO(values []domain.Wish) []gin.H {
	out := make([]gin.H, 0, len(values))
	for _, v := range values {
		out = append(out, wishDTO(v))
	}
	return out
}
