package accounthttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	application "github.com/nanfang/just-for-you/server/internal/account/application"
	domain "github.com/nanfang/just-for-you/server/internal/account/domain"
	httpserver "github.com/nanfang/just-for-you/server/internal/platform/httpapi"
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
	Execute(context.Context, application.GetProfileQuery) (application.Profile, error)
}
type UpdateProfileExecutor interface {
	Execute(context.Context, application.UpdateProfileCommand) (application.Profile, error)
}
type RequestAvatarUploadExecutor interface {
	Execute(context.Context, application.RequestAvatarUploadCommand) (application.UploadGrant, error)
}
type Dependencies struct {
	Login               LoginExecutor
	RefreshSession      RefreshSessionExecutor
	Logout              LogoutExecutor
	GetProfile          GetProfileExecutor
	UpdateProfile       UpdateProfileExecutor
	RequestAvatarUpload RequestAvatarUploadExecutor
}
type Handler struct{ useCases Dependencies }

func NewHandler(useCases Dependencies) *Handler { return &Handler{useCases: useCases} }
func (h *Handler) RegisterPublic(router *gin.Engine) {
	router.POST("/v1/auth/wechat", h.login)
	router.POST("/v1/auth/refresh", h.refresh)
	router.POST("/v1/auth/logout", h.logout)
}
func (h *Handler) RegisterProtected(group *gin.RouterGroup) {
	group.GET("/users/me", h.getMe)
	group.PUT("/users/me", h.updateProfile)
	group.POST("/users/me/avatar-upload", h.requestAvatarUpload)
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
		} else if errors.Is(err, domain.ErrInvalidAvatar) {
			httpserver.WriteError(c, http.StatusBadRequest, "INVALID_AVATAR", "头像无效或已过期")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务暂时不可用")
		}
		return
	}
	httpserver.WriteData(c, 200, toDTO(user))
}
func (h *Handler) requestAvatarUpload(c *gin.Context) {
	var input struct {
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpserver.WriteError(c, 400, "INVALID_AVATAR", "头像信息无效")
		return
	}
	grant, err := h.useCases.RequestAvatarUpload.Execute(c, application.RequestAvatarUploadCommand{UserID: httpserver.UserID(c), ContentType: input.ContentType, Size: input.Size})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAvatar) {
			httpserver.WriteError(c, 400, "INVALID_AVATAR", "请选择有效的头像图片")
		} else {
			_ = c.Error(err)
			httpserver.WriteError(c, 500, "INTERNAL_ERROR", "服务暂时不可用")
		}
		return
	}
	httpserver.WriteData(c, 201, grant)
}
func toDTO(profile application.Profile) gin.H {
	return gin.H{"id": profile.ID(), "displayName": profile.DisplayName(), "avatarRef": profile.AvatarRef(), "avatarUrl": profile.AvatarURL}
}
