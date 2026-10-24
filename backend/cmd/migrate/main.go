package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	database "github.com/nanfang/just-for-you/server/internal/platform/database"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	version := flag.Int("version", 0, "single migration version; 0 applies every pending version")
	directory := flag.String("dir", "migrations", "migration directory")
	flag.Parse()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "just_for_you:local-development@tcp(127.0.0.1:3306)/just_for_you?parseTime=true&multiStatements=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = database.ApplyMigrations(
		context.Background(),
		db,
		*directory,
		*direction,
		*version,
	); err != nil {
		log.Fatal(err)
	}
	log.Printf("migrations %s complete", *direction)
}
