package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/user"
	"net/http"
)

type userHandler struct{ service user.Service }

func (h userHandler) update(c *gin.Context) {
	var in user.UpdateInput
	if c.ShouldBindJSON(&in) != nil {
		writeError(c, 400, "INVALID_REQUEST", "请求格式错误")
		return
	}
	u, e := h.service.UpdateMe(c, c.GetInt64(userIDKey), in)
	if e != nil {
		writeError(c, 400, "INVALID_PROFILE", "昵称需为 1 至 30 个字符")
		return
	}
	writeData(c, http.StatusOK, u)
}
