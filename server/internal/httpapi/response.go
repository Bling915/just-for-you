package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
)

func writeData(c *gin.Context, status int, data any) {
	httpserver.WriteData(c, status, data)
}

func writeError(c *gin.Context, status int, code, message string) {
	httpserver.WriteError(c, status, code, message)
}
