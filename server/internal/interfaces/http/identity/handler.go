package identityhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/application/identity"
	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
)

type LoginExecutor interface {
	Execute(context.Context, application.LoginCommand) (application.Session, error)
}
type RefreshSessionExecutor interface {
	Execute(context.Context, application.RefreshSessionCommand) (application.Session, error)
}
type LogoutExecutor interface {
	Execute(context.Context, application.LogoutCommand) error
}
type GetProfileExecutor interface {
	Execute(context.Context, application.GetProfileQuery) (domain.User, error)
}
type UpdateProfileExecutor interface {
	Execute(context.Context, application.UpdateProfileCommand) (domain.User, error)
}
type UseCases struct {
	Login          LoginExecutor
	RefreshSession RefreshSessionExecutor
	Logout         LogoutExecutor
	GetProfile     GetProfileExecutor
	UpdateProfile  UpdateProfileExecutor
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
	session, err := h.useCases.Login.Execute(c, application.LoginCommand{Code: input.Code})
	if err != nil {
		if errors.Is(err, domain.ErrCodeRequired) {
			httpserver.WriteError(c, http.StatusBadRequest, "LOGIN_FAILED", "登录凭证不能为空")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, http.StatusServiceUnavailable, "TEMPORARY_UNAVAILABLE", "登录服务暂时不可用")
		}
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
	session, err := h.useCases.RefreshSession.Execute(c, application.RefreshSessionCommand{RefreshToken: input.RefreshToken})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			httpserver.WriteError(c, http.StatusUnauthorized, "INVALID_REFRESH", "请重新登录")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用")
		}
		return
	}
	httpserver.WriteData(c, 200, session)
}
func (h *Handler) logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.ShouldBindJSON(&input) == nil {
		_ = h.useCases.Logout.Execute(c, application.LogoutCommand{RefreshToken: input.RefreshToken})
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) getMe(c *gin.Context) {
	user, err := h.useCases.GetProfile.Execute(c, application.GetProfileQuery{UserID: httpserver.UserID(c)})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			httpserver.WriteError(c, http.StatusNotFound, "USER_NOT_FOUND", "用户不存在")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用")
		}
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
	user, err := h.useCases.UpdateProfile.Execute(c, application.UpdateProfileCommand{UserID: httpserver.UserID(c), DisplayName: input.DisplayName, AvatarRef: input.AvatarRef})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidNickname) {
			httpserver.WriteError(c, http.StatusBadRequest, "INVALID_PROFILE", "昵称需为 1 至 30 个字符")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用")
		}
		return
	}
	httpserver.WriteData(c, 200, toDTO(user))
}
func toDTO(user domain.User) gin.H {
	return gin.H{"id": user.ID(), "displayName": user.DisplayName(), "avatarRef": user.AvatarRef()}
}
