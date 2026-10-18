package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nanfang/just-for-you/server/internal/couple"
	"github.com/nanfang/just-for-you/server/internal/httpapi"
	identityapplication "github.com/nanfang/just-for-you/server/internal/identity/application"
	identityinfrastructure "github.com/nanfang/just-for-you/server/internal/identity/infrastructure"
	identityhttp "github.com/nanfang/just-for-you/server/internal/identity/interfaces/http"
	platformclock "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	"github.com/nanfang/just-for-you/server/internal/platform/database"
	"github.com/nanfang/just-for-you/server/internal/platform/httpserver"
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
	coupleService := couple.NewService(couple.SQLRepository{DB: db}, time.Now, couple.RandomCode)
	router := httpserver.NewRouter(httpserver.Dependencies{Tokens: tokens, Registrars: []httpserver.RouteRegistrar{identityhttp.NewHandler(identityService), httpapi.NewRelationshipRegistrar(coupleService)}})
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
