package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Dependencies struct{ DB *sql.DB }

func NewRouter(_ Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", func(c *gin.Context) {
		writeData(c, http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}
