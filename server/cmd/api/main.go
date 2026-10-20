package main

import (
	"log"
	"net/http"
	"time"

	identityapplication "github.com/nanfang/just-for-you/server/internal/application/identity"
	infraClock "github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/config"
	wechat "github.com/nanfang/just-for-you/server/internal/infrastructure/identity/wechat"
	identitymysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/identity"
	database "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/shared"
	jwt "github.com/nanfang/just-for-you/server/internal/infrastructure/security/jwt"
	identityhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/identity"
	httpserver "github.com/nanfang/just-for-you/server/internal/interfaces/http/shared"
	momentapplication "github.com/nanfang/just-for-you/server/internal/moment/application"
	momentinfrastructure "github.com/nanfang/just-for-you/server/internal/moment/infrastructure"
	momenthttp "github.com/nanfang/just-for-you/server/internal/moment/interfaces/http"
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
	tokens := jwt.NewJWTManager([]byte(cfg.JWTSecret))
	identityRepository := identitymysql.MySQLRepository{DB: db}
	identityProvider := wechat.WeChatIdentityProvider{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}
	identityUseCases := identityhttp.UseCases{
		Login:          identityapplication.NewLoginUseCase(identityProvider, identityRepository, tokens, infraClock.System{}, jwt.GenerateRefreshToken),
		RefreshSession: identityapplication.NewRefreshSessionUseCase(identityRepository, tokens, infraClock.System{}, jwt.GenerateRefreshToken),
		Logout:         identityapplication.NewLogoutUseCase(identityRepository, infraClock.System{}),
		GetProfile:     identityapplication.NewGetProfileUseCase(identityRepository),
		UpdateProfile:  identityapplication.NewUpdateProfileUseCase(identityRepository),
	}
	relationshipService := relationshipapplication.NewService(relationshipinfrastructure.NewMySQLRepository(db), infraClock.System{}, relationshipinfrastructure.GenerateInvitationCode)
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
		infraClock.System{},
	)
	router := httpserver.NewRouter(httpserver.Dependencies{
		Tokens: tokens,
		Registrars: []httpserver.RouteRegistrar{
			identityhttp.NewHandler(identityUseCases),
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
