package companionshiphttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/companionship/application"
	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
	"net/http"
	"time"
)

type UpdatePresenceExecutor interface {
	Execute(context.Context, application.UpdatePresenceCommand) (domain.Presence, error)
}
type SendNudgeExecutor interface {
	Execute(context.Context, application.SendNudgeCommand) (domain.Nudge, error)
}
type GetSnapshotExecutor interface {
	Execute(context.Context, application.GetSnapshotQuery) (application.Snapshot, error)
}
type GetNotificationPreferencesExecutor interface {
	Execute(context.Context, application.GetNotificationPreferencesQuery) (application.NotificationPreferences, error)
}
type UpdateNotificationPreferencesExecutor interface {
	Execute(context.Context, application.UpdateNotificationPreferencesCommand) (application.NotificationPreferences, error)
}
type Dependencies struct {
	UpdatePresence                UpdatePresenceExecutor
	SendNudge                     SendNudgeExecutor
	GetSnapshot                   GetSnapshotExecutor
	GetNotificationPreferences    GetNotificationPreferencesExecutor
	UpdateNotificationPreferences UpdateNotificationPreferencesExecutor
}
type Handler struct {
	dependencies Dependencies
}

func NewHandler(dependencies Dependencies) *Handler { return &Handler{dependencies: dependencies} }
func (*Handler) RegisterPublic(*gin.Engine)         {}
func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.PUT("/presence", h.updatePresence)
	group.GET("/presence", h.snapshot)
	group.POST("/nudges", h.sendNudge)
	group.GET("/notification-preferences", h.getPreferences)
	group.PUT("/notification-preferences", h.updatePreferences)
}
func (h *Handler) getPreferences(c *gin.Context) {
	if h.dependencies.GetNotificationPreferences == nil {
		httpserver.WriteError(c, 503, "NOTIFICATION_UNAVAILABLE", "通知暂不可用")
		return
	}
	v, e := h.dependencies.GetNotificationPreferences.Execute(c, application.GetNotificationPreferencesQuery{UserID: httpserver.UserID(c)})
	if e != nil {
		writeError(c, e)
		return
	}
	httpserver.WriteData(c, 200, v)
}
func (h *Handler) updatePreferences(c *gin.Context) {
	var v application.NotificationPreferences
	if c.ShouldBindJSON(&v) != nil {
		httpserver.WriteError(c, 400, "INVALID_ARGUMENT", "通知设置无效")
		return
	}
	out, e := h.dependencies.UpdateNotificationPreferences.Execute(c, application.UpdateNotificationPreferencesCommand{UserID: httpserver.UserID(c), Preferences: v})
	if e != nil {
		writeError(c, e)
		return
	}
	httpserver.WriteData(c, 200, out)
}
func (h *Handler) updatePresence(c *gin.Context) {
	var input struct {
		Preset     string     `json:"preset"`
		CustomText string     `json:"customText"`
		ExpiresAt  *time.Time `json:"expiresAt"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_PRESENCE", "状态内容不正确")
		return
	}
	result, err := h.dependencies.UpdatePresence.Execute(c, application.UpdatePresenceCommand{UserID: httpserver.UserID(c), Preset: input.Preset, CustomText: input.CustomText, ExpiresAt: input.ExpiresAt})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusOK, presenceDTO(&result))
}
func (h *Handler) snapshot(c *gin.Context) {
	result, err := h.dependencies.GetSnapshot.Execute(c, application.GetSnapshotQuery{UserID: httpserver.UserID(c)})
	if err != nil {
		writeError(c, err)
		return
	}
	var latest gin.H
	if result.LatestNudge != nil {
		latest = nudgeDTO(*result.LatestNudge)
	}
	httpserver.WriteData(c, http.StatusOK, gin.H{"partnerPresence": presenceDTO(result.PartnerPresence), "latestNudge": latest})
}
func (h *Handler) sendNudge(c *gin.Context) {
	var input struct {
		Kind string `json:"kind"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "INVALID_NUDGE", "互动类型不正确")
		return
	}
	result, err := h.dependencies.SendNudge.Execute(c, application.SendNudgeCommand{UserID: httpserver.UserID(c), Kind: input.Kind})
	if err != nil {
		writeError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, nudgeDTO(result))
}
func presenceDTO(p *domain.Presence) gin.H {
	if p == nil {
		return nil
	}
	return gin.H{"id": p.ID(), "userId": p.UserID(), "preset": p.Preset(), "customText": p.CustomText(), "expiresAt": p.ExpiresAt(), "updatedAt": p.UpdatedAt()}
}
func nudgeDTO(n domain.Nudge) gin.H {
	return gin.H{"id": n.ID(), "senderUserId": n.SenderID(), "recipientUserId": n.RecipientID(), "kind": n.Kind(), "createdAt": n.CreatedAt()}
}
func writeError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用"
	switch {
	case errors.Is(err, domain.ErrInvalidPresence):
		status, code, message = http.StatusBadRequest, "INVALID_PRESENCE", "状态内容不正确"
	case errors.Is(err, domain.ErrPresenceTextTooLong):
		status, code, message = http.StatusBadRequest, "PRESENCE_TEXT_TOO_LONG", "状态最多二十个字"
	case errors.Is(err, domain.ErrInvalidPresenceExpiry):
		status, code, message = http.StatusBadRequest, "INVALID_PRESENCE_EXPIRY", "结束时间需要晚于现在"
	case errors.Is(err, domain.ErrInvalidNudge):
		status, code, message = http.StatusBadRequest, "INVALID_NUDGE", "互动类型不正确"
	case errors.Is(err, domain.ErrNudgeThrottled):
		status, code, message = http.StatusTooManyRequests, "NUDGE_THROTTLED", "慢一点，让心意先飞一会儿"
	case errors.Is(err, coupledomain.ErrNoActiveCouple):
		status, code, message = http.StatusNotFound, "NO_ACTIVE_COUPLE", "还没有绑定关系"
	default:
		_ = c.Error(err)
	}
	httpserver.WriteError(c, status, code, message)
}
