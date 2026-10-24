package couplehttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/couple/application"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	httpserver "github.com/nanfang/just-for-you/server/internal/platform/httpapi"
	"net/http"
)

type CreateInvitationExecutor interface {
	Execute(context.Context, application.CreateInvitationCommand) (domain.Invitation, error)
}
type PreviewInvitationExecutor interface {
	Execute(context.Context, application.PreviewInvitationQuery) (domain.Invitation, error)
}
type BindCoupleExecutor interface {
	Execute(context.Context, application.BindCoupleCommand) (domain.Couple, error)
}
type GetCoupleExecutor interface {
	Execute(context.Context, application.GetCoupleQuery) (application.CurrentCouple, error)
}
type EndCoupleExecutor interface {
	Execute(context.Context, application.EndCoupleCommand) error
}
type EnableSoloTestExecutor interface {
	Execute(context.Context, application.EnableSoloTestCommand) (domain.Couple, error)
}
type Dependencies struct {
	CreateInvitation  CreateInvitationExecutor
	PreviewInvitation PreviewInvitationExecutor
	BindCouple        BindCoupleExecutor
	GetCouple         GetCoupleExecutor
	EndCouple         EndCoupleExecutor
	EnableSoloTest    EnableSoloTestExecutor
}
type Handler struct{ useCases Dependencies }

func NewHandler(useCases Dependencies) *Handler { return &Handler{useCases: useCases} }
func (*Handler) RegisterPublic(*gin.Engine)     {}
func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.POST("/couples/invitations", h.create)
	group.GET("/couples/invitations/:code", h.preview)
	group.POST("/couples/bind", h.bind)
	group.GET("/couples/me", h.get)
	group.DELETE("/couples/me", h.end)
	if h.useCases.EnableSoloTest != nil {
		group.POST("/couples/solo-test", h.enableSoloTest)
	}
}
func (h *Handler) create(c *gin.Context) {
	x, err := h.useCases.CreateInvitation.Execute(c, application.CreateInvitationCommand{UserID: httpserver.UserID(c)})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 201, invitationDTO(x))
}
func (h *Handler) preview(c *gin.Context) {
	x, err := h.useCases.PreviewInvitation.Execute(c, application.PreviewInvitationQuery{UserID: httpserver.UserID(c), Code: c.Param("code")})
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
	x, err := h.useCases.BindCouple.Execute(c, application.BindCoupleCommand{UserID: httpserver.UserID(c), Code: input.Code})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 201, relationshipDTO(x))
}
func (h *Handler) get(c *gin.Context) {
	x, err := h.useCases.GetCouple.Execute(c, application.GetCoupleQuery{UserID: httpserver.UserID(c)})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, 200, currentCoupleDTO(x))
}
func (h *Handler) end(c *gin.Context) {
	if err := h.useCases.EndCouple.Execute(c, application.EndCoupleCommand{UserID: httpserver.UserID(c)}); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) enableSoloTest(c *gin.Context) {
	x, err := h.useCases.EnableSoloTest.Execute(c, application.EnableSoloTestCommand{UserID: httpserver.UserID(c)})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	httpserver.WriteData(c, http.StatusCreated, relationshipDTO(x))
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
	case errors.Is(err, domain.ErrNoActiveCouple):
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
func relationshipDTO(x domain.Couple) gin.H {
	return gin.H{"id": x.ID(), "userAId": x.UserAID(), "userBId": x.UserBID(), "status": x.Status(), "boundAt": x.BoundAt()}
}
func currentCoupleDTO(x application.CurrentCouple) gin.H {
	relationship := x.Couple
	return gin.H{
		"id": relationship.ID(), "userAId": relationship.UserAID(), "userBId": relationship.UserBID(),
		"status": relationship.Status(), "boundAt": relationship.BoundAt(), "soloTest": x.SoloTest,
		"partner": gin.H{"userId": x.Partner.UserID, "displayName": x.Partner.DisplayName, "avatarUrl": x.Partner.AvatarURL},
	}
}
