package todohttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	shared "github.com/nanfang/just-for-you/server/internal/platform/httpapi"
	app "github.com/nanfang/just-for-you/server/internal/todo/application"
	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type ListTodosExecutor interface {
	Execute(context.Context, app.ListTodosQuery) ([]domain.Todo, error)
}

type CreateTodoExecutor interface {
	Execute(context.Context, app.CreateTodoCommand) (domain.Todo, error)
}

type SetTodoCompletedExecutor interface {
	Execute(context.Context, app.SetTodoCompletedCommand) (domain.Todo, error)
}

type DeleteTodoExecutor interface {
	Execute(context.Context, app.DeleteTodoCommand) error
}

type Dependencies struct {
	List     ListTodosExecutor
	Create   CreateTodoExecutor
	Complete SetTodoCompletedExecutor
	Delete   DeleteTodoExecutor
}

type Handler struct {
	dependencies Dependencies
}

func NewHandler(dependencies Dependencies) *Handler {
	return &Handler{dependencies: dependencies}
}

func (*Handler) RegisterPublic(*gin.Engine) {}

func (handler *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/todos", handler.list)
	group.POST("/todos", handler.create)
	group.PATCH("/todos/:id", handler.update)
	group.DELETE("/todos/:id", handler.delete)
}

func (handler *Handler) list(c *gin.Context) {
	items, err := handler.dependencies.List.Execute(
		c,
		app.ListTodosQuery{UserID: shared.UserID(c)},
	)
	if err != nil {
		fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, dto(item))
	}
	shared.WriteData(c, http.StatusOK, out)
}

func (handler *Handler) create(c *gin.Context) {
	var input struct {
		Title string     `json:"title"`
		DueAt *time.Time `json:"dueAt"`
	}
	if c.ShouldBindJSON(&input) != nil {
		shared.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "待办内容不正确")
		return
	}
	item, err := handler.dependencies.Create.Execute(
		c,
		app.CreateTodoCommand{
			UserID: shared.UserID(c),
			Title:  input.Title,
			DueAt:  input.DueAt,
		},
	)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusCreated, dto(item))
}

func (handler *Handler) update(c *gin.Context) {
	itemID, err := itemID(c)
	var input struct {
		Completed bool `json:"completed"`
	}
	if c.ShouldBindJSON(&input) != nil {
		err = domain.ErrInvalidTitle
	}
	if err != nil {
		fail(c, err)
		return
	}
	item, err := handler.dependencies.Complete.Execute(
		c,
		app.SetTodoCompletedCommand{
			UserID:    shared.UserID(c),
			ID:        itemID,
			Completed: input.Completed,
		},
	)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusOK, dto(item))
}

func (handler *Handler) delete(c *gin.Context) {
	itemID, err := itemID(c)
	if err == nil {
		err = handler.dependencies.Delete.Execute(
			c,
			app.DeleteTodoCommand{UserID: shared.UserID(c), ID: itemID},
		)
	}
	if err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func dto(item domain.Todo) gin.H {
	return gin.H{
		"id":                item.ID(),
		"creatorUserId":     item.CreatorID(),
		"completedByUserId": item.CompletedBy(),
		"title":             item.Title(),
		"dueAt":             item.DueAt(),
		"status":            item.Status(),
		"completedAt":       item.CompletedAt(),
		"createdAt":         item.CreatedAt(),
		"updatedAt":         item.UpdatedAt(),
	}
}

func itemID(c *gin.Context) (int64, error) {
	value, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, domain.ErrNotFound
	}
	return value, nil
}

func fail(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(err, domain.ErrInvalidTitle):
		status, code, message = http.StatusBadRequest, "INVALID_ARGUMENT", "内容不正确"
	case errors.Is(err, domain.ErrNotCreator):
		status, code, message = http.StatusForbidden, "NOT_CREATOR", "只能删除自己创建的内容"
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "内容不存在"
	case errors.Is(err, coupledomain.ErrNoActiveCouple):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定关系"
	default:
		_ = c.Error(err)
	}
	shared.WriteError(c, status, code, message)
}
