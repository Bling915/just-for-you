package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/auth"
	"github.com/nanfang/just-for-you/server/internal/user"
)

type Dependencies struct {
	DB          *sql.DB
	AuthService *auth.Service
	Tokens      auth.TokenManager
	UserService user.Service
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", func(c *gin.Context) {
		writeData(c, http.StatusOK, gin.H{"status": "ok"})
	})
	h := authHandler{service: deps.AuthService}
	router.POST("/v1/auth/wechat", h.login)
	router.POST("/v1/auth/refresh", h.refresh)
	router.POST("/v1/auth/logout", h.logout)
	protected := router.Group("/v1", requireAuth(deps.Tokens))
	protected.PUT("/users/me", userHandler{service: deps.UserService}.update)
	return router
}
