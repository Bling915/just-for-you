package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nanfang/just-for-you/server/internal/config"
	"github.com/nanfang/just-for-you/server/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	if err := httpapi.NewRouter(httpapi.Dependencies{DB: db}).Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
