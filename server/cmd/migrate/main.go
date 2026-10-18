package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	flag.Parse()
	name := "000001_init." + *direction + ".sql"
	contents, err := os.ReadFile(filepath.Join("migrations", name))
	if err != nil {
		log.Fatal(err)
	}
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "just_for_you:local-development@tcp(127.0.0.1:3306)/just_for_you?parseTime=true&multiStatements=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(string(contents)); err != nil {
		log.Fatal(err)
	}
	fmt.Println("applied", name)
}
