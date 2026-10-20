package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouteRegistrar interface {
	RegisterPublic(*gin.Engine)
	RegisterProtected(*gin.RouterGroup)
}
type Dependencies struct {
	Tokens     AccessTokenVerifier
	Registrars []RouteRegistrar
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
	router.GET("/healthz", func(c *gin.Context) { WriteData(c, http.StatusOK, gin.H{"status": "ok"}) })
	for _, registrar := range deps.Registrars {
		registrar.RegisterPublic(router)
	}
	protected := router.Group("/v1", Authenticate(deps.Tokens))
	for _, registrar := range deps.Registrars {
		registrar.RegisterProtected(protected)
	}
	return router
}
