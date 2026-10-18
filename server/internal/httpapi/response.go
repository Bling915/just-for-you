package httpapi

import "github.com/gin-gonic/gin"

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeData(c *gin.Context, status int, data any) { c.JSON(status, gin.H{"data": data}) }
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": errorBody{Code: code, Message: message}})
}
