package wishhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	shared "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
	app "github.com/nanfang/just-for-you/server/internal/wish/application"
	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type ListWishesExecutor interface {
	Execute(context.Context, app.ListWishesQuery) ([]domain.Wish, error)
}

type CreateWishExecutor interface {
	Execute(context.Context, app.CreateWishCommand) (domain.Wish, error)
}

type CompleteWishExecutor interface {
	Execute(context.Context, app.CompleteWishCommand) (domain.Wish, error)
}

type ReopenWishExecutor interface {
	Execute(context.Context, app.ReopenWishCommand) (domain.Wish, error)
}

type DeleteWishExecutor interface {
	Execute(context.Context, app.DeleteWishCommand) error
}

type Dependencies struct {
	List     ListWishesExecutor
	Create   CreateWishExecutor
	Complete CompleteWishExecutor
	Reopen   ReopenWishExecutor
	Delete   DeleteWishExecutor
}

type Handler struct {
	dependencies Dependencies
}

func NewHandler(dependencies Dependencies) *Handler {
	return &Handler{dependencies: dependencies}
}

func (*Handler) RegisterPublic(*gin.Engine) {}

func (handler *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/wishes", handler.list)
	group.POST("/wishes", handler.create)
	group.PATCH("/wishes/:id", handler.update)
	group.DELETE("/wishes/:id", handler.delete)
}

func (handler *Handler) list(c *gin.Context) {
	items, err := handler.dependencies.List.Execute(
		c,
		app.ListWishesQuery{UserID: shared.UserID(c)},
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
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&input) != nil {
		shared.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "愿望内容不正确")
		return
	}
	item, err := handler.dependencies.Create.Execute(
		c,
		app.CreateWishCommand{
			UserID:      shared.UserID(c),
			Title:       input.Title,
			Description: input.Description,
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
		Completed      bool   `json:"completed"`
		LinkedMomentID *int64 `json:"linkedMomentId"`
	}
	if c.ShouldBindJSON(&input) != nil {
		err = domain.ErrInvalidTitle
	}
	if err != nil {
		fail(c, err)
		return
	}
	var item domain.Wish
	if input.Completed {
		item, err = handler.dependencies.Complete.Execute(
			c,
			app.CompleteWishCommand{
				UserID:   shared.UserID(c),
				ID:       itemID,
				MomentID: input.LinkedMomentID,
			},
		)
	} else {
		item, err = handler.dependencies.Reopen.Execute(
			c,
			app.ReopenWishCommand{UserID: shared.UserID(c), ID: itemID},
		)
	}
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
			app.DeleteWishCommand{UserID: shared.UserID(c), ID: itemID},
		)
	}
	if err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func dto(item domain.Wish) gin.H {
	return gin.H{
		"id":                item.ID(),
		"creatorUserId":     item.CreatorID(),
		"completedByUserId": item.CompletedBy(),
		"linkedMomentId":    item.LinkedMomentID(),
		"title":             item.Title(),
		"description":       item.Description(),
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
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidDescription):
		status, code, message = http.StatusBadRequest, "INVALID_ARGUMENT", "内容不正确"
	case errors.Is(err, domain.ErrInvalidMoment):
		status, code, message = http.StatusBadRequest, "INVALID_MOMENT", "关联动态不正确"
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
