package architecture

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestCoupleExperienceMigration(t *testing.T) {
	up := migrationSQL(t, "000002_couple_experience.up.sql")
	down := migrationSQL(t, "000002_couple_experience.down.sql")

	root, err := sql.Open("mysql", migrationTestDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.Ping(); err != nil {
		t.Skipf("migration test MySQL is unavailable: %v", err)
	}

	databaseName := fmt.Sprintf("just_for_you_migration_%d", time.Now().UnixNano())
	if _, err := root.Exec("CREATE DATABASE " + databaseName + " CHARACTER SET utf8mb4"); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP DATABASE " + databaseName)

	db, err := sql.Open("mysql", migrationDatabaseDSN(databaseName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(migrationSQL(t, "000001_init.up.sql")); err != nil {
		t.Fatalf("apply initial migration: %v", err)
	}
	if _, err := db.Exec(up); err != nil {
		t.Fatalf("apply couple-experience migration: %v", err)
	}

	for _, table := range []string{"presences", "nudges", "notification_preferences", "notification_deliveries", "anniversaries", "todos", "wishes"} {
		assertTableExists(t, db, table, true)
	}
	for _, column := range []string{"occurred_at", "updated_at", "latitude", "longitude"} {
		assertColumnExists(t, db, "moments", column, true)
	}
	for _, index := range []string{"idx_moments_calendar", "idx_presences_couple_updated", "idx_nudges_couple_created", "idx_anniversaries_couple_date", "idx_todos_couple_status", "idx_wishes_couple_status"} {
		assertIndexExists(t, db, index, true)
	}

	if _, err := db.Exec(down); err != nil {
		t.Fatalf("apply couple-experience down migration: %v", err)
	}
	assertTableExists(t, db, "users", true)
	assertTableExists(t, db, "moments", true)
	for _, table := range []string{"presences", "nudges", "notification_preferences", "notification_deliveries", "anniversaries", "todos", "wishes"} {
		assertTableExists(t, db, table, false)
	}
	for _, column := range []string{"occurred_at", "updated_at", "latitude", "longitude"} {
		assertColumnExists(t, db, "moments", column, false)
	}
}

func migrationSQL(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(contents)
}

func migrationTestDSN() string {
	if configured := os.Getenv("MIGRATION_TEST_DSN"); configured != "" {
		return configured
	}
	return "root@unix(/tmp/mysql.sock)/?parseTime=true&multiStatements=true"
}

func migrationDatabaseDSN(databaseName string) string {
	return "root@unix(/tmp/mysql.sock)/" + databaseName + "?parseTime=true&multiStatements=true"
}

func assertTableExists(t *testing.T, db *sql.DB, table string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 1; got != want {
		t.Fatalf("table %s exists=%v, want %v", table, got, want)
	}
}

func assertColumnExists(t *testing.T, db *sql.DB, table, column string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? AND column_name=?", table, column).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 1; got != want {
		t.Fatalf("column %s.%s exists=%v, want %v", table, column, got, want)
	}
}

func assertIndexExists(t *testing.T, db *sql.DB, index string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND index_name=?", index).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count > 0; got != want {
		t.Fatalf("index %s exists=%v, want %v", index, got, want)
	}
}
