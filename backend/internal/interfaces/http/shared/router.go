package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/platform/auth"
	"github.com/nanfang/just-for-you/server/internal/platform/httpresponse"
)

type RouteRegistrar interface {
	RegisterPublic(*gin.Engine)
	RegisterProtected(*gin.RouterGroup)
}
type Dependencies struct {
	Tokens     auth.AccessTokenVerifier
	Registrars []RouteRegistrar
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
	router.GET("/healthz", func(c *gin.Context) { httpresponse.WriteData(c, http.StatusOK, gin.H{"status": "ok"}) })
	for _, registrar := range deps.Registrars {
		registrar.RegisterPublic(router)
	}
	protected := router.Group("/v1", auth.Authenticate(deps.Tokens))
	for _, registrar := range deps.Registrars {
		registrar.RegisterProtected(protected)
	}
	return router
}

var WriteData = httpresponse.WriteData
var WriteError = httpresponse.WriteError
var UserID = auth.UserID
