package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nanfang/just-for-you/server/internal/auth"
	"github.com/nanfang/just-for-you/server/internal/couple"
	"github.com/nanfang/just-for-you/server/internal/httpapi"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	"github.com/nanfang/just-for-you/server/internal/platform/database"
	"github.com/nanfang/just-for-you/server/internal/user"
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
	tokens := auth.NewTokenManager([]byte(cfg.JWTSecret))
	authService := auth.NewService(auth.HTTPWeChatExchanger{AppID: cfg.WeChatAppID, Secret: cfg.WeChatSecret, Client: &http.Client{Timeout: 5 * time.Second}}, auth.SQLRepository{DB: db}, tokens, time.Now)
	coupleService := couple.NewService(couple.SQLRepository{DB: db}, time.Now, couple.RandomCode)
	if err := httpapi.NewRouter(httpapi.Dependencies{DB: db, AuthService: authService, Tokens: tokens, UserService: user.Service{DB: db}, CoupleService: coupleService}).Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
