package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/auth"
	"net/http"
	"strings"
)

const userIDKey = "userID"

func requireAuth(tokens auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		id, e := tokens.ParseAccess(raw)
		if e != nil {
			writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "请重新登录")
			c.Abort()
			return
		}
		c.Set(userIDKey, id)
		c.Next()
	}
}
