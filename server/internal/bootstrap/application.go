package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/nanfang/just-for-you/server/internal/infrastructure/config"
	database "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/shared"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
)

type Application struct {
	server   *http.Server
	database *sql.DB
}

func New() (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	db, err := database.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}
	identity, tokens := buildIdentity(cfg, db)
	relationship, getRelationship := buildRelationship(db, cfg.EnableSoloTest)
	moment, err := buildMoment(cfg, db, getRelationship)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	router := httpserver.NewRouter(httpserver.Dependencies{Tokens: tokens, Registrars: []httpserver.RouteRegistrar{identity, relationship, moment}})
	return &Application{server: newHTTPServer(cfg.HTTPAddr, router), database: db}, nil
}

func newApplicationForTest(deps httpserver.Dependencies) *Application {
	return &Application{server: newHTTPServer(":0", httpserver.NewRouter(deps))}
}
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
}

func (a *Application) Run(ctx context.Context) error {
	errs := make(chan error, 1)
	go func() { errs <- a.server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.Close(shutdownCtx)
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
func (a *Application) Close(ctx context.Context) error {
	var shutdownErr, errorClose error
	if a.server != nil {
		shutdownErr = a.server.Shutdown(ctx)
	}
	if a.database != nil {
		errorClose = a.database.Close()
	}
	return errors.Join(shutdownErr, errorClose)
}
