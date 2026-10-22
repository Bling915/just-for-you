package momenthttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
)

type RequestPhotoUploadExecutor interface {
	Execute(context.Context, application.RequestPhotoUploadCommand) (application.UploadGrant, error)
}
type PublishMomentExecutor interface {
	Execute(context.Context, application.PublishMomentCommand) (domain.Moment, error)
}
type UpdateMomentExecutor interface {
	Execute(context.Context, application.UpdateMomentCommand) (domain.Moment, error)
}
type GetTimelineExecutor interface {
	Execute(context.Context, application.GetTimelineQuery) (application.TimelinePage, error)
}
type GetHomeExecutor interface {
	Execute(context.Context, application.GetHomeQuery) (application.HomeResult, error)
}
type DeleteMomentExecutor interface {
	Execute(context.Context, application.DeleteMomentCommand) error
}
type UseCases struct {
	RequestPhotoUpload RequestPhotoUploadExecutor
	PublishMoment      PublishMomentExecutor
	UpdateMoment       UpdateMomentExecutor
	GetTimeline        GetTimelineExecutor
	GetHome            GetHomeExecutor
	DeleteMoment       DeleteMomentExecutor
	ReadModels         *application.ReadModels
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
	group.PATCH("/moments/:id", h.update)
	group.GET("/moments", h.timeline)
	group.GET("/home", h.home)
	group.GET("/moments/calendar", h.calendar)
	group.GET("/moments/map", h.memoryMap)
	group.GET("/reports/monthly", h.monthlyReport)
	group.DELETE("/moments/:id", h.delete)
}

func (h *Handler) calendar(c *gin.Context) {
	if h.useCases.ReadModels == nil {
		httpserver.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	items, err := h.useCases.ReadModels.Calendar(c, httpserver.UserID(c), c.Query("month"))
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, 200, gin.H{"days": items})
}
func (h *Handler) memoryMap(c *gin.Context) {
	if h.useCases.ReadModels == nil {
		httpserver.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		httpserver.WriteError(c, 400, "INVALID_ARGUMENT", "时间范围无效")
		return
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil || !to.After(from) {
		httpserver.WriteError(c, 400, "INVALID_ARGUMENT", "时间范围无效")
		return
	}
	items, err := h.useCases.ReadModels.Map(c, httpserver.UserID(c), from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, 200, gin.H{"points": items})
}
func (h *Handler) monthlyReport(c *gin.Context) {
	if h.useCases.ReadModels == nil {
		httpserver.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		return
	}
	report, err := h.useCases.ReadModels.Report(c, httpserver.UserID(c), c.Query("month"))
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, 200, report)
}

func (h *Handler) requestUpload(c *gin.Context) {
	var input application.UploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_UPLOAD", "照片信息无效")
		return
	}
	grant, err := h.useCases.RequestPhotoUpload.Execute(c, application.RequestPhotoUploadCommand{UserID: httpserver.UserID(c), ContentType: input.ContentType, Size: input.Size})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, grant)
}

func (h *Handler) publish(c *gin.Context) {
	var input struct {
		Text       string        `json:"text"`
		Status     domain.Status `json:"status"`
		PlaceName  string        `json:"placeName"`
		Latitude   *float64      `json:"latitude"`
		Longitude  *float64      `json:"longitude"`
		PhotoKeys  []string      `json:"photoKeys"`
		OccurredAt time.Time     `json:"occurredAt"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "发布内容格式无效")
		return
	}
	moment, err := h.useCases.PublishMoment.Execute(c, application.PublishMomentCommand{UserID: httpserver.UserID(c), Text: input.Text, Status: input.Status, PlaceName: input.PlaceName, Latitude: input.Latitude, Longitude: input.Longitude, PhotoKeys: input.PhotoKeys, OccurredAt: input.OccurredAt})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, momentDTO(moment))
}

func (h *Handler) update(c *gin.Context) {
	momentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || momentID <= 0 {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "动态编号无效")
		return
	}
	var input struct {
		Text       string        `json:"text"`
		Status     domain.Status `json:"status"`
		PlaceName  string        `json:"placeName"`
		Latitude   *float64      `json:"latitude"`
		Longitude  *float64      `json:"longitude"`
		PhotoKeys  []string      `json:"photoKeys"`
		OccurredAt time.Time     `json:"occurredAt"`
		Version    time.Time     `json:"version"`
	}
	if c.ShouldBindJSON(&input) != nil || input.OccurredAt.IsZero() || input.Version.IsZero() {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "更新内容格式无效")
		return
	}
	moment, err := h.useCases.UpdateMoment.Execute(c, application.UpdateMomentCommand{UserID: httpserver.UserID(c), MomentID: momentID, Text: input.Text, Status: input.Status, PlaceName: input.PlaceName, Latitude: input.Latitude, Longitude: input.Longitude, PhotoKeys: input.PhotoKeys, OccurredAt: input.OccurredAt, Version: input.Version})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusOK, momentDTO(moment))
}

func (h *Handler) timeline(c *gin.Context) {
	limit, err := optionalPositiveInt(c.Query("limit"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "limit 必须是正整数")
		return
	}
	spaceType, owner, err := parseTimelineScope(c.Query("space"), c.Query("owner"))
	if err != nil {
		writeError(c, err)
		return
	}
	page, err := h.useCases.GetTimeline.Execute(c, application.GetTimelineQuery{UserID: httpserver.UserID(c), Cursor: c.Query("cursor"), Limit: limit, SpaceType: spaceType, Owner: owner})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusOK, gin.H{"moments": momentsDTO(page.Moments, page.PhotoURLs), "nextCursor": page.NextCursor})
}

func parseTimelineScope(rawSpace, rawOwner string) (domain.SpaceType, application.TimelineOwner, error) {
	space := domain.SpaceType(rawSpace)
	owner := application.TimelineOwner(rawOwner)
	if space == "" {
		space = domain.SpaceShared
	}
	if space == domain.SpaceShared {
		if owner != "" {
			return "", "", application.ErrInvalidSpaceQuery
		}
		return space, "", nil
	}
	if space != domain.SpacePersonal {
		return "", "", application.ErrInvalidSpaceQuery
	}
	if owner == "" {
		owner = application.TimelineOwnerMe
	}
	if owner != application.TimelineOwnerMe && owner != application.TimelineOwnerPartner {
		return "", "", application.ErrInvalidSpaceQuery
	}
	return space, owner, nil
}

func (h *Handler) home(c *gin.Context) {
	home, err := h.useCases.GetHome.Execute(c, application.GetHomeQuery{UserID: httpserver.UserID(c)})
	if err != nil {
		writeError(c, err)
		return
	}
	var latest any
	if home.PartnerLatestStatus != nil {
		latest = momentDTO(*home.PartnerLatestStatus, home.PhotoURLs)
	}
	httpserver.WriteData(c, http.StatusOK, gin.H{"partnerLatestStatus": latest, "recent": momentsDTO(home.Recent, home.PhotoURLs), "pinnedAnniversaries": home.PinnedAnniversaries, "hasMorePinned": home.HasMorePinned})
}

func (h *Handler) delete(c *gin.Context) {
	momentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || momentID <= 0 {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_ARGUMENT", "动态编号无效")
		return
	}
	if err := h.useCases.DeleteMoment.Execute(c, application.DeleteMomentCommand{UserID: httpserver.UserID(c), MomentID: momentID}); err != nil {
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
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(err, domain.ErrTooManyPhotos):
		status, code, message = http.StatusBadRequest, "TOO_MANY_PHOTOS", "最多上传九张照片"
	case errors.Is(err, domain.ErrEmptyMoment):
		status, code, message = http.StatusBadRequest, "EMPTY_MOMENT", "请记录一些内容"
	case errors.Is(err, domain.ErrInvalidUpload), errors.Is(err, domain.ErrDuplicatePhoto):
		status, code, message = http.StatusBadRequest, "INVALID_UPLOAD", "照片无效或已过期"
	case errors.Is(err, domain.ErrTextTooLong):
		status, code, message = http.StatusBadRequest, "TEXT_TOO_LONG", "文字不能超过 500 个字"
	case errors.Is(err, domain.ErrPlaceTooLong):
		status, code, message = http.StatusBadRequest, "PLACE_TOO_LONG", "位置名称不能超过 80 个字"
	case errors.Is(err, domain.ErrInvalidStatus):
		status, code, message = http.StatusBadRequest, "INVALID_STATUS", "状态无效"
	case errors.Is(err, domain.ErrInvalidOccurredAt):
		status, code, message = http.StatusBadRequest, "INVALID_OCCURRED_AT", "记录时间无效"
	case errors.Is(err, domain.ErrInvalidCoordinates):
		status, code, message = http.StatusBadRequest, "INVALID_COORDINATES", "位置坐标无效"
	case errors.Is(err, domain.ErrMomentConflict):
		status, code, message = http.StatusConflict, "MOMENT_CONFLICT", "这条记录已在别处更新，请重新加载"
	case errors.Is(err, domain.ErrNotAuthor):
		status, code, message = http.StatusForbidden, "NOT_AUTHOR", "只能删除自己的动态"
	case errors.Is(err, domain.ErrMomentNotFound):
		status, code, message = http.StatusNotFound, "MOMENT_NOT_FOUND", "动态不存在"
	case errors.Is(err, application.ErrInvalidSpaceQuery):
		status, code, message = http.StatusBadRequest, "INVALID_SPACE_QUERY", "小事空间参数无效"
	case errors.Is(err, relationshipdomain.ErrNoActiveRelationship):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定情侣关系"
	}
	if status >= http.StatusInternalServerError {
		_ = c.Error(err)
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
		"status": moment.Status(), "placeName": moment.Place(), "latitude": moment.Latitude(), "longitude": moment.Longitude(), "photos": photos,
		"spaceType": moment.SpaceType(), "visibility": moment.Visibility(), "kind": moment.Kind(), "sourceMomentId": moment.SourceMomentID(), "dailyQuestionId": moment.DailyQuestionID(),
		"occurredAt": moment.OccurredAt(), "createdAt": moment.CreatedAt(), "updatedAt": moment.UpdatedAt(),
	}
}

func momentsDTO(moments []domain.Moment, photoURLs map[string]string) []gin.H {
	result := make([]gin.H, 0, len(moments))
	for _, moment := range moments {
		result = append(result, momentDTO(moment, photoURLs))
	}
	return result
}
