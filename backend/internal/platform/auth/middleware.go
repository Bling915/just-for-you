package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nanfang/just-for-you/server/internal/platform/httpresponse"
)

const userIDKey = "userID"

type AccessTokenVerifier interface{ ParseAccess(string) (int64, error) }

func Authenticate(verifier AccessTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		userID, err := verifier.ParseAccess(raw)
		if err != nil {
			httpresponse.WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", "请重新登录")
			c.Abort()
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}
func UserID(c *gin.Context) int64 { return c.GetInt64(userIDKey) }
