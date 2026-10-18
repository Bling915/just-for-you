package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/auth"
	"net/http"
)

type authHandler struct{ service *auth.Service }

func (h authHandler) login(c *gin.Context) {
	if h.service == nil {
		writeError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "服务暂不可用")
		return
	}
	var in struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, 400, "INVALID_REQUEST", "请求格式错误")
		return
	}
	session, err := h.service.Login(c, in.Code)
	if err != nil {
		writeError(c, 400, "LOGIN_FAILED", "登录失败")
		return
	}
	writeData(c, 200, session)
}
func (h authHandler) refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refreshToken"`
	}
	if h.service == nil || c.ShouldBindJSON(&in) != nil {
		writeError(c, 400, "INVALID_REFRESH", "请重新登录")
		return
	}
	s, e := h.service.Refresh(c, in.RefreshToken)
	if e != nil {
		writeError(c, 401, "INVALID_REFRESH", "请重新登录")
		return
	}
	writeData(c, 200, s)
}
func (h authHandler) logout(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refreshToken"`
	}
	if h.service != nil && c.ShouldBindJSON(&in) == nil {
		_ = h.service.Revoke(c, in.RefreshToken)
	}
	c.Status(http.StatusNoContent)
}
