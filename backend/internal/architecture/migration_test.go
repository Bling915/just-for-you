package architecture

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	database "github.com/nanfang/just-for-you/server/internal/platform/database"
)

func TestBaselineMigrationDefinition(t *testing.T) {
	up := migrationSQL(t, "000001_baseline.up.sql")
	down := migrationSQL(t, "000001_baseline.down.sql")
	required := []string{
		"status VARCHAR(80) NOT NULL DEFAULT ''",
		"uq_moment_daily_question",
		"idx_moments_timeline",
		"idx_moments_calendar",
		"idx_moments_personal_timeline",
		"idx_daily_questions_couple_created",
		"schema_migrations",
	}
	for _, fragment := range required {
		if !strings.Contains(up, fragment) {
			t.Fatalf("baseline up migration missing %q", fragment)
		}
	}
	for _, forbidden := range []string{"status" + "_code", "notification_" + "deliveries"} {
		if strings.Contains(up, forbidden) || strings.Contains(down, forbidden) {
			t.Fatalf("baseline still contains removed schema %q", forbidden)
		}
	}
	removedKind := "k" + "ind"
	momentStart := strings.Index(up, "CREATE TABLE moments")
	momentEnd := strings.Index(up[momentStart:], ";\n\nCREATE TABLE")
	momentDefinition := up[momentStart : momentStart+momentEnd]
	if strings.Contains(momentDefinition, " "+removedKind+" ") ||
		strings.Contains(momentDefinition, "`"+removedKind+"`") {
		t.Fatal("baseline still contains removed moment discriminator column")
	}
}

func TestBaselineMigrationUpDownUpAndTracking(t *testing.T) {
	root, err := sql.Open("mysql", migrationTestDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.Ping(); err != nil {
		t.Skipf("migration test MySQL is unavailable: %v", err)
	}
	databaseName := fmt.Sprintf("just_for_you_baseline_%d", time.Now().UnixNano())
	if _, err := root.Exec("CREATE DATABASE " + databaseName + " CHARACTER SET utf8mb4"); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP DATABASE " + databaseName)
	db, err := sql.Open("mysql", migrationDatabaseDSN(databaseName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	migrations := filepath.Join("..", "..", "migrations")

	if err = database.ApplyMigrations(ctx, db, migrations, "up", 0); err != nil {
		t.Fatalf("apply baseline: %v", err)
	}
	if err = database.ApplyMigrations(ctx, db, migrations, "up", 0); err != nil {
		t.Fatalf("reapply baseline must be idempotent: %v", err)
	}
	assertMigrationCount(t, db, 1)
	assertTableExists(t, db, "moments", true)
	assertTableExists(t, db, "notification_"+"deliveries", false)
	assertColumnExists(t, db, "moments", "status", true)
	assertColumnExists(t, db, "moments", "status"+"_code", false)
	assertColumnExists(t, db, "moments", "k"+"ind", false)
	assertColumnType(t, db, "moments", "status", 80)
	for _, index := range []string{
		"idx_moments_timeline",
		"idx_moments_calendar",
		"idx_moments_personal_timeline",
		"uq_moment_daily_question",
		"idx_daily_questions_couple_created",
	} {
		assertIndexExists(t, db, index, true)
	}

	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	if _, err = db.Exec(
		"INSERT INTO users(id,wechat_openid) VALUES(1,'baseline-a'),(2,'baseline-b')",
	); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"INSERT INTO couples(id,user_a_id,user_b_id,status,bound_at) VALUES(7,1,2,'active',?)",
		now,
	); err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(
		"INSERT INTO moments(couple_id,author_user_id,space_type,visibility,body,status,place_name,latitude,longitude,occurred_at,created_at,updated_at) VALUES(7,1,'shared','couple','基线记录','想你了','',NULL,NULL,?,?,?)",
		now,
		now,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	momentID, _ := result.LastInsertId()
	if momentID == 0 {
		t.Fatal("moment insert did not return id")
	}
	if _, err = db.Exec(
		"INSERT INTO daily_questions(id,couple_id,question_date,question_text,source,created_at,updated_at) VALUES(3,7,?,?,?,?,?)",
		"2026-10-24",
		"今天最开心的小事？",
		"fallback",
		now,
		now,
	); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"INSERT INTO moments(couple_id,author_user_id,space_type,visibility,daily_question_id,body,status,place_name,occurred_at,created_at,updated_at) VALUES(7,1,'shared','couple',3,'每日一句','','',?,?,?)",
		now,
		now,
		now,
	); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"INSERT INTO moments(couple_id,author_user_id,space_type,visibility,daily_question_id,body,status,place_name,occurred_at,created_at,updated_at) VALUES(7,2,'shared','couple',3,'重复投影','','',?,?,?)",
		now,
		now,
		now,
	); err == nil {
		t.Fatal("daily question projection must be unique")
	}

	if err = database.ApplyMigrations(ctx, db, migrations, "down", 0); err != nil {
		t.Fatalf("apply baseline down: %v", err)
	}
	assertTableExists(t, db, "moments", false)
	assertTableExists(t, db, "daily_questions", false)
	assertMigrationCount(t, db, 0)
	if err = database.ApplyMigrations(ctx, db, migrations, "up", 0); err != nil {
		t.Fatalf("reapply baseline after down: %v", err)
	}
	assertTableExists(t, db, "moments", true)
	assertMigrationCount(t, db, 1)
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
	if configured := os.Getenv("MIGRATION_TEST_DSN"); configured != "" {
		if index := strings.Index(configured, ")/"); index >= 0 {
			query := ""
			if queryIndex := strings.Index(configured[index+2:], "?"); queryIndex >= 0 {
				query = configured[index+2+queryIndex:]
			}
			return configured[:index+2] + databaseName + query
		}
	}
	return "root@unix(/tmp/mysql.sock)/" + databaseName + "?parseTime=true&multiStatements=true"
}

func assertMigrationCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("schema_migrations count=%d want=%d", count, want)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, table string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?",
		table,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 1; got != want {
		t.Fatalf("table %s exists=%v, want %v", table, got, want)
	}
}

func assertColumnExists(t *testing.T, db *sql.DB, table, column string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? AND column_name=?",
		table,
		column,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count == 1; got != want {
		t.Fatalf("column %s.%s exists=%v, want %v", table, column, got, want)
	}
}

func assertColumnType(t *testing.T, db *sql.DB, table, column string, maxRunes int) {
	t.Helper()
	var length int
	if err := db.QueryRow(
		"SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? AND column_name=?",
		table,
		column,
	).Scan(&length); err != nil {
		t.Fatal(err)
	}
	if length != maxRunes {
		t.Fatalf("%s.%s max length=%d want=%d", table, column, length, maxRunes)
	}
}

func assertIndexExists(t *testing.T, db *sql.DB, index string, want bool) {
	t.Helper()
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND index_name=?",
		index,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got := count > 0; got != want {
		t.Fatalf("index %s exists=%v, want %v", index, got, want)
	}
}
