package relationshiphttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
	"github.com/nanfang/just-for-you/server/internal/relationship/domain"
	"net/http"
)

type UseCases interface {
	CreateInvitation(context.Context, int64) (domain.Invitation, error)
	PreviewInvitation(context.Context, int64, string) (domain.Invitation, error)
	BindRelationship(context.Context, int64, string) (domain.Relationship, error)
	GetRelationship(context.Context, int64) (domain.Relationship, error)
	EndRelationship(context.Context, int64) error
}
type Handler struct{ useCases UseCases }

func NewHandler(useCases UseCases) *Handler { return &Handler{useCases: useCases} }
func (*Handler) RegisterPublic(*gin.Engine) {}
func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/couples/invitations", h.create)
	group.GET("/couples/invitations/:code", h.preview)
	group.POST("/couples/bind", h.bind)
	group.GET("/couples/me", h.get)
	group.DELETE("/couples/me", h.end)
}
func (h *Handler) create(c *gin.Context) {
	x, err := h.useCases.CreateInvitation(c, httpserver.UserID(c))
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 201, invitationDTO(x))
}
func (h *Handler) preview(c *gin.Context) {
	x, err := h.useCases.PreviewInvitation(c, httpserver.UserID(c), c.Param("code"))
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 200, invitationDTO(x))
}
func (h *Handler) bind(c *gin.Context) {
	var input struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, 400, "INVITE_INVALID", "邀请码无效")
		return
	}
	x, err := h.useCases.BindRelationship(c, httpserver.UserID(c), input.Code)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 201, relationshipDTO(x))
}
func (h *Handler) get(c *gin.Context) {
	x, err := h.useCases.GetRelationship(c, httpserver.UserID(c))
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 200, relationshipDTO(x))
}
func (h *Handler) end(c *gin.Context) {
	if err := h.useCases.EndRelationship(c, httpserver.UserID(c)); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func writeDomainError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	switch {
	case errors.Is(err, domain.ErrInvitationInvalid):
		status, code = http.StatusBadRequest, "INVITE_INVALID"
	case errors.Is(err, domain.ErrInvitationExpired):
		status, code = http.StatusBadRequest, "INVITE_EXPIRED"
	case errors.Is(err, domain.ErrInvitationUsed):
		status, code = http.StatusBadRequest, "INVITE_USED"
	case errors.Is(err, domain.ErrCannotBindSelf):
		status, code = http.StatusBadRequest, "CANNOT_BIND_SELF"
	case errors.Is(err, domain.ErrAlreadyBound):
		status, code = http.StatusBadRequest, "ALREADY_BOUND"
	case errors.Is(err, domain.ErrNoActiveRelationship):
		code = "NO_ACTIVE_COUPLE"
		status = 404
	}
	message := "操作失败"
	if status == http.StatusInternalServerError {
		message = "服务暂时不可用"
		_ = c.Error(err)
	}
	httpserver.WriteError(c, status, code, message)
}
func invitationDTO(x domain.Invitation) gin.H {
	return gin.H{"code": x.Code(), "creatorUserId": x.CreatorID(), "expiresAt": x.ExpiresAt()}
}
func relationshipDTO(x domain.Relationship) gin.H {
	return gin.H{"id": x.ID(), "userAId": x.UserAID(), "userBId": x.UserBID(), "status": x.Status(), "boundAt": x.BoundAt()}
}
