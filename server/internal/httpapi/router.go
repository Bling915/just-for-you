package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/auth"
	"github.com/nanfang/just-for-you/server/internal/couple"
	"github.com/nanfang/just-for-you/server/internal/user"
)

type Dependencies struct {
	DB            *sql.DB
	AuthService   *auth.Service
	Tokens        auth.TokenManager
	UserService   user.Service
	CoupleService *couple.Service
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
	ch := coupleHandler{s: deps.CoupleService}
	protected.POST("/couples/invitations", ch.create)
	protected.GET("/couples/invitations/:code", ch.preview)
	protected.POST("/couples/bind", ch.accept)
	protected.GET("/couples/me", ch.get)
	protected.DELETE("/couples/me", ch.unbind)
	return router
}
