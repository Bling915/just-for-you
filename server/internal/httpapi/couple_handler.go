package httpapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/couple"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
	"net/http"
)

type coupleHandler struct{ s *couple.Service }

func (h coupleHandler) create(c *gin.Context) {
	x, e := h.s.CreateInvite(c, httpserver.UserID(c))
	if e != nil {
		writeError(c, 400, "INVITE_FAILED", "无法创建邀请码")
		return
	}
	writeData(c, 201, x)
}
func (h coupleHandler) preview(c *gin.Context) {
	x, e := h.s.PreviewInvite(c, httpserver.UserID(c), c.Param("code"))
	if e != nil {
		h.err(c, e)
		return
	}
	writeData(c, 200, x)
}
func (h coupleHandler) accept(c *gin.Context) {
	var in struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, 400, "INVITE_INVALID", "邀请码无效")
		return
	}
	x, e := h.s.AcceptInvite(c, httpserver.UserID(c), in.Code)
	if e != nil {
		h.err(c, e)
		return
	}
	writeData(c, 201, x)
}
func (h coupleHandler) get(c *gin.Context) {
	x, e := h.s.GetActive(c, httpserver.UserID(c))
	if e != nil {
		h.err(c, e)
		return
	}
	writeData(c, 200, x)
}
func (h coupleHandler) unbind(c *gin.Context) {
	if e := h.s.Unbind(c, httpserver.UserID(c)); e != nil {
		h.err(c, e)
		return
	}
	c.Status(204)
}
func (h coupleHandler) err(c *gin.Context, e error) {
	code := "INVITE_INVALID"
	status := 400
	switch {
	case errors.Is(e, couple.ErrInviteExpired):
		code = "INVITE_EXPIRED"
	case errors.Is(e, couple.ErrInviteUsed):
		code = "INVITE_USED"
	case errors.Is(e, couple.ErrCannotBindSelf):
		code = "CANNOT_BIND_SELF"
	case errors.Is(e, couple.ErrAlreadyBound):
		code = "ALREADY_BOUND"
	case errors.Is(e, couple.ErrNoActiveCouple):
		code = "NO_ACTIVE_COUPLE"
		status = http.StatusNotFound
	}
	writeError(c, status, code, "操作失败")
}
