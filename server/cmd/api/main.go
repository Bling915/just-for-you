package main

import (
	"log"
	"net/http"
	"time"

	identityapplication "github.com/nanfang/just-for-you/server/internal/identity/application"
	identityinfrastructure "github.com/nanfang/just-for-you/server/internal/identity/infrastructure"
	identityhttp "github.com/nanfang/just-for-you/server/internal/identity/interfaces/http"
	momentapplication "github.com/nanfang/just-for-you/server/internal/moment/application"
	momentinfrastructure "github.com/nanfang/just-for-you/server/internal/moment/infrastructure"
	momenthttp "github.com/nanfang/just-for-you/server/internal/moment/interfaces/http"
	platformclock "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	"github.com/nanfang/just-for-you/server/internal/platform/database"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
	relationshipapplication "github.com/nanfang/just-for-you/server/internal/relationship/application"
	relationshipinfrastructure "github.com/nanfang/just-for-you/server/internal/relationship/infrastructure"
	relationshiphttp "github.com/nanfang/just-for-you/server/internal/relationship/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tokens := identityinfrastructure.NewJWTManager([]byte(cfg.JWTSecret))
	identityService := identityapplication.NewService(identityinfrastructure.WeChatIdentityProvider{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}, identityinfrastructure.MySQLRepository{DB: db}, tokens, platformclock.System{}, identityinfrastructure.GenerateRefreshToken)
	relationshipService := relationshipapplication.NewService(relationshipinfrastructure.NewMySQLRepository(db), platformclock.System{}, relationshipinfrastructure.GenerateInvitationCode)
	momentRepository := momentinfrastructure.NewMySQLRepository(db)
	photoStorage, err := momentinfrastructure.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		log.Fatal(err)
	}
	momentService := momentapplication.NewService(
		momentinfrastructure.NewActiveRelationshipReader(relationshipService),
		momentRepository,
		momentRepository,
		photoStorage,
		platformclock.System{},
	)
	router := httpserver.NewRouter(httpserver.Dependencies{
		Tokens: tokens,
		Registrars: []httpserver.RouteRegistrar{
			identityhttp.NewHandler(identityService),
			relationshiphttp.NewHandler(relationshipService),
			momenthttp.NewHandler(momentService),
		},
	})
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
