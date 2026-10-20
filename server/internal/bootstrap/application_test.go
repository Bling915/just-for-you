package bootstrap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
)

type testTokens struct{}

func (testTokens) ParseAccess(string) (int64, error) { return 1, nil }

type testRegistrar struct{}

func (testRegistrar) RegisterPublic(r *gin.Engine) {
	r.GET("/test-public", func(c *gin.Context) { c.Status(204) })
}
func (testRegistrar) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/test-protected", func(c *gin.Context) { c.Status(204) })
}

func TestApplicationBuildsRegisteredRoutes(t *testing.T) {
	app := newApplicationForTest(httpserver.Dependencies{Tokens: testTokens{}, Registrars: []httpserver.RouteRegistrar{testRegistrar{}}})
	if app.server.MaxHeaderBytes != 1<<20 {
		t.Fatalf("expected bounded headers, got %d", app.server.MaxHeaderBytes)
	}
	for _, path := range []string{"/healthz", "/test-public", "/v1/test-protected"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Authorization", "Bearer token")
		app.server.Handler.ServeHTTP(response, request)
		if response.Code == http.StatusNotFound {
			t.Fatalf("route %s was not registered", path)
		}
	}
	_ = app.Close(context.Background())
}
