package momenthttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/moment/application"
	"github.com/nanfang/just-for-you/server/internal/moment/domain"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/relationship/domain"
)

type UseCases interface {
	RequestPhotoUpload(context.Context, int64, application.UploadInput) (application.UploadGrant, error)
	PublishMoment(context.Context, int64, application.PublishInput) (domain.Moment, error)
	GetTimeline(context.Context, int64, string, int) (application.TimelinePage, error)
	GetHome(context.Context, int64) (application.Home, error)
	DeleteMoment(context.Context, int64, int64) error
}

type Handler struct {
	useCases UseCases
}

func NewHandler(useCases UseCases) *Handler {
	return &Handler{useCases: useCases}
}

func (*Handler) RegisterPublic(*gin.Engine) {}

func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/uploads", h.requestUpload)
	group.POST("/moments", h.publish)
	group.GET("/moments", h.timeline)
	group.GET("/home", h.home)
	group.DELETE("/moments/:id", h.delete)
}

func (h *Handler) requestUpload(c *gin.Context) {
	var input application.UploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_UPLOAD", "照片信息无效")
		return
	}
	grant, err := h.useCases.RequestPhotoUpload(c, httpserver.UserID(c), input)
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, grant)
}

func (h *Handler) publish(c *gin.Context) {
	var input application.PublishInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "发布内容格式无效")
		return
	}
	moment, err := h.useCases.PublishMoment(c, httpserver.UserID(c), input)
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, momentDTO(moment))
}

func (h *Handler) timeline(c *gin.Context) {
	limit, err := optionalPositiveInt(c.Query("limit"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "limit 必须是正整数")
		return
	}
	page, err := h.useCases.GetTimeline(c, httpserver.UserID(c), c.Query("cursor"), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusOK, gin.H{"moments": momentsDTO(page.Moments, page.PhotoURLs), "nextCursor": page.NextCursor})
}

func (h *Handler) home(c *gin.Context) {
	home, err := h.useCases.GetHome(c, httpserver.UserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	var latest any
	if home.PartnerLatestStatus != nil {
		latest = momentDTO(*home.PartnerLatestStatus, home.PhotoURLs)
	}
	httpserver.WriteData(c, http.StatusOK, gin.H{"partnerLatestStatus": latest, "recent": momentsDTO(home.Recent, home.PhotoURLs)})
}

func (h *Handler) delete(c *gin.Context) {
	momentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || momentID <= 0 {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "动态编号无效")
		return
	}
	if err := h.useCases.DeleteMoment(c, httpserver.UserID(c), momentID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func optionalPositiveInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New("not a positive integer")
	}
	return value, nil
}

func writeError(c *gin.Context, err error) {
	status, code, message := http.StatusBadRequest, "INVALID_ARGUMENT", "操作失败"
	switch {
	case errors.Is(err, domain.ErrTooManyPhotos):
		code, message = "TOO_MANY_PHOTOS", "最多上传九张照片"
	case errors.Is(err, domain.ErrEmptyMoment):
		code, message = "EMPTY_MOMENT", "请记录一些内容"
	case errors.Is(err, domain.ErrInvalidUpload), errors.Is(err, domain.ErrDuplicatePhoto):
		code, message = "INVALID_UPLOAD", "照片无效或已过期"
	case errors.Is(err, domain.ErrTextTooLong):
		code, message = "TEXT_TOO_LONG", "文字不能超过 500 个字"
	case errors.Is(err, domain.ErrPlaceTooLong):
		code, message = "PLACE_TOO_LONG", "位置名称不能超过 80 个字"
	case errors.Is(err, domain.ErrInvalidStatus):
		code, message = "INVALID_STATUS", "状态无效"
	case errors.Is(err, domain.ErrNotAuthor):
		status, code, message = http.StatusForbidden, "NOT_AUTHOR", "只能删除自己的动态"
	case errors.Is(err, domain.ErrMomentNotFound):
		status, code, message = http.StatusNotFound, "MOMENT_NOT_FOUND", "动态不存在"
	case errors.Is(err, relationshipdomain.ErrNoActiveRelationship):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定情侣关系"
	}
	httpserver.WriteError(c, status, code, message)
}

func momentDTO(moment domain.Moment, photoURLs ...map[string]string) gin.H {
	urls := map[string]string{}
	if len(photoURLs) > 0 {
		urls = photoURLs[0]
	}
	photos := make([]gin.H, 0, len(moment.Photos()))
	for _, photo := range moment.Photos() {
		photos = append(photos, gin.H{"objectKey": photo.ObjectKey(), "url": urls[photo.ObjectKey()], "order": photo.Order()})
	}
	return gin.H{
		"id": moment.ID(), "authorId": moment.AuthorID(), "text": moment.Text(),
		"status": moment.Status(), "placeName": moment.Place(), "photos": photos,
		"createdAt": moment.CreatedAt(),
	}
}

func momentsDTO(moments []domain.Moment, photoURLs map[string]string) []gin.H {
	result := make([]gin.H, 0, len(moments))
	for _, moment := range moments {
		result = append(result, momentDTO(moment, photoURLs))
	}
	return result
}
