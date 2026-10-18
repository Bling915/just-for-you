package identityhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/identity/application"
	"github.com/nanfang/just-for-you/server/internal/identity/domain"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
)

type UseCases interface {
	LoginWithWeChat(context.Context, string) (application.Session, error)
	RefreshSession(context.Context, string) (application.Session, error)
	Logout(context.Context, string) error
	GetMe(context.Context, int64) (domain.User, error)
	UpdateProfile(context.Context, int64, string, string) (domain.User, error)
}
type Handler struct{ useCases UseCases }

func NewHandler(useCases UseCases) *Handler { return &Handler{useCases: useCases} }
func (h *Handler) RegisterPublic(router *gin.Engine) {
	router.POST("/v1/auth/wechat", h.login)
	router.POST("/v1/auth/refresh", h.refresh)
	router.POST("/v1/auth/logout", h.logout)
}
func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/users/me", h.getMe)
	group.PUT("/users/me", h.updateProfile)
}
func (h *Handler) login(c *gin.Context) {
	var input struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, 400, "INVALID_REQUEST", "请求格式错误")
		return
	}
	session, err := h.useCases.LoginWithWeChat(c, input.Code)
	if err != nil {
		httpserver.WriteError(c, 400, "LOGIN_FAILED", "登录失败")
		return
	}
	httpserver.WriteData(c, 200, session)
}
func (h *Handler) refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, 400, "INVALID_REFRESH", "请重新登录")
		return
	}
	session, err := h.useCases.RefreshSession(c, input.RefreshToken)
	if err != nil {
		httpserver.WriteError(c, 401, "INVALID_REFRESH", "请重新登录")
		return
	}
	httpserver.WriteData(c, 200, session)
}
func (h *Handler) logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.ShouldBindJSON(&input) == nil {
		_ = h.useCases.Logout(c, input.RefreshToken)
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) getMe(c *gin.Context) {
	user, err := h.useCases.GetMe(c, httpserver.UserID(c))
	if err != nil {
		httpserver.WriteError(c, 404, "USER_NOT_FOUND", "用户不存在")
		return
	}
	httpserver.WriteData(c, 200, toDTO(user))
}
func (h *Handler) updateProfile(c *gin.Context) {
	var input struct {
		DisplayName string `json:"displayName"`
		AvatarRef   string `json:"avatarRef"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, 400, "INVALID_REQUEST", "请求格式错误")
		return
	}
	user, err := h.useCases.UpdateProfile(c, httpserver.UserID(c), input.DisplayName, input.AvatarRef)
	if err != nil {
		code := "PROFILE_UPDATE_FAILED"
		if errors.Is(err, domain.ErrInvalidNickname) {
			code = "INVALID_PROFILE"
		}
		httpserver.WriteError(c, 400, code, "昵称需为 1 至 30 个字符")
		return
	}
	httpserver.WriteData(c, 200, toDTO(user))
}
func toDTO(user domain.User) gin.H {
	return gin.H{"id": user.ID(), "displayName": user.DisplayName(), "avatarRef": user.AvatarRef()}
}
