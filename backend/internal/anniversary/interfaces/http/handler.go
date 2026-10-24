package anniversaryhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	app "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	shared "github.com/nanfang/just-for-you/server/internal/platform/httpapi"
)

type ListAnniversariesExecutor interface {
	Execute(context.Context, app.ListAnniversariesQuery) ([]domain.Anniversary, error)
}

type GetAnniversaryExecutor interface {
	Execute(context.Context, app.GetAnniversaryQuery) (domain.Anniversary, error)
}

type CreateAnniversaryExecutor interface {
	Execute(context.Context, app.CreateAnniversaryCommand) (domain.Anniversary, error)
}

type UpdateAnniversaryExecutor interface {
	Execute(context.Context, app.UpdateAnniversaryCommand) (domain.Anniversary, error)
}

type DeleteAnniversaryExecutor interface {
	Execute(context.Context, app.DeleteAnniversaryCommand) error
}

type RequestCoverUploadExecutor interface {
	Execute(context.Context, app.RequestCoverUploadCommand) (app.CoverUploadGrant, error)
}

type CoverURLExecutor interface {
	Execute(context.Context, app.CoverURLQuery) (string, error)
}

type Dependencies struct {
	List         ListAnniversariesExecutor
	Get          GetAnniversaryExecutor
	Create       CreateAnniversaryExecutor
	Update       UpdateAnniversaryExecutor
	Delete       DeleteAnniversaryExecutor
	RequestCover RequestCoverUploadExecutor
	CoverURL     CoverURLExecutor
}

type Handler struct {
	dependencies Dependencies
}

func NewHandler(dependencies Dependencies) *Handler {
	return &Handler{dependencies: dependencies}
}

func (*Handler) RegisterPublic(*gin.Engine) {}

func (handler *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/anniversaries", handler.list)
	group.POST("/anniversaries", handler.create)
	group.GET("/anniversaries/:id", handler.get)
	group.PATCH("/anniversaries/:id", handler.update)
	group.POST("/anniversaries/:id/cover-upload", handler.requestCover)
	group.DELETE("/anniversaries/:id", handler.delete)
}

func (handler *Handler) list(c *gin.Context) {
	items, err := handler.dependencies.List.Execute(
		c,
		app.ListAnniversariesQuery{UserID: shared.UserID(c)},
	)
	if err != nil {
		fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		dto, err := handler.dto(c, item)
		if err != nil {
			fail(c, err)
			return
		}
		out = append(out, dto)
	}
	shared.WriteData(c, http.StatusOK, out)
}

func (handler *Handler) create(c *gin.Context) {
	var input anniversaryInput
	if c.ShouldBindJSON(&input) != nil {
		shared.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "纪念日内容不正确")
		return
	}
	item, err := handler.dependencies.Create.Execute(
		c,
		app.CreateAnniversaryCommand{UserID: shared.UserID(c), Draft: input.draft()},
	)
	if err != nil {
		fail(c, err)
		return
	}
	dto, err := handler.dto(c, item)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusCreated, dto)
}

func (handler *Handler) get(c *gin.Context) {
	itemID, err := itemID(c)
	if err != nil {
		fail(c, err)
		return
	}
	item, err := handler.dependencies.Get.Execute(
		c,
		app.GetAnniversaryQuery{UserID: shared.UserID(c), ID: itemID},
	)
	if err != nil {
		fail(c, err)
		return
	}
	dto, err := handler.dto(c, item)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusOK, dto)
}

func (handler *Handler) update(c *gin.Context) {
	itemID, err := itemID(c)
	var input anniversaryInput
	if c.ShouldBindJSON(&input) != nil {
		err = domain.ErrInvalidTitle
	}
	if err != nil {
		fail(c, err)
		return
	}
	item, err := handler.dependencies.Update.Execute(
		c,
		app.UpdateAnniversaryCommand{
			UserID: shared.UserID(c),
			ID:     itemID,
			Draft:  input.draft(),
		},
	)
	if err != nil {
		fail(c, err)
		return
	}
	dto, err := handler.dto(c, item)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusOK, dto)
}

func (handler *Handler) requestCover(c *gin.Context) {
	itemID, err := itemID(c)
	var input struct {
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
	}
	if c.ShouldBindJSON(&input) != nil {
		err = domain.ErrInvalidCover
	}
	if err != nil {
		fail(c, err)
		return
	}
	grant, err := handler.dependencies.RequestCover.Execute(
		c,
		app.RequestCoverUploadCommand{
			UserID:      shared.UserID(c),
			ID:          itemID,
			ContentType: input.ContentType,
			Size:        input.Size,
		},
	)
	if err != nil {
		fail(c, err)
		return
	}
	shared.WriteData(c, http.StatusCreated, grant)
}

func (handler *Handler) delete(c *gin.Context) {
	itemID, err := itemID(c)
	if err == nil {
		err = handler.dependencies.Delete.Execute(
			c,
			app.DeleteAnniversaryCommand{UserID: shared.UserID(c), ID: itemID},
		)
	}
	if err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (handler *Handler) dto(c *gin.Context, item domain.Anniversary) (gin.H, error) {
	url, err := handler.dependencies.CoverURL.Execute(
		c,
		app.CoverURLQuery{CoverRef: item.CoverRef()},
	)
	if err != nil {
		return nil, err
	}
	return gin.H{
		"id":            item.ID(),
		"creatorUserId": item.CreatorID(),
		"title":         item.Title(),
		"date":          item.Date(),
		"countMode":     item.CountMode(),
		"repeatsYearly": item.RepeatsYearly(),
		"pinned":        item.Pinned(),
		"coverRef":      item.CoverRef(),
		"coverUrl":      url,
		"message":       item.Message(),
		"systemKind":    item.SystemKind(),
		"createdAt":     item.CreatedAt(),
		"updatedAt":     item.UpdatedAt(),
	}, nil
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

func (input anniversaryInput) draft() domain.AnniversaryDraft {
	if input.CountMode == "" {
		input.CountMode = domain.CountModeCountdown
	}
	return domain.AnniversaryDraft{
		Title:         input.Title,
		Date:          input.Date,
		CountMode:     input.CountMode,
		RepeatsYearly: input.Repeats,
		Pinned:        input.Pinned,
		CoverRef:      input.CoverRef,
		Message:       input.Message,
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
		errors.Is(err, domain.ErrInvalidDate),
		errors.Is(err, domain.ErrInvalidMessage),
		errors.Is(err, domain.ErrInvalidCountMode):
		status, code, message = http.StatusBadRequest, "INVALID_ARGUMENT", "内容不正确"
	case errors.Is(err, domain.ErrInvalidCover):
		status, code, message = http.StatusBadRequest, "INVALID_COVER", "封面无效或已过期"
	case errors.Is(err, domain.ErrNotCreator):
		status, code, message = http.StatusForbidden, "NOT_CREATOR", "只能删除自己创建的内容"
	case errors.Is(err, domain.ErrSystemAnniversary):
		status, code, message = http.StatusForbidden, "SYSTEM_ANNIVERSARY", "系统纪念日不能删除"
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "内容不存在"
	case errors.Is(err, coupledomain.ErrNoActiveCouple):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定关系"
	default:
		_ = c.Error(err)
	}
	shared.WriteError(c, status, code, message)
}
