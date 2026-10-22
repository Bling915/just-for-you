package architecture

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestAvatarAnniversaryMigrationShape(t *testing.T) {
	up := migrationSQL(t, "000003_avatar_anniversary_experience.up.sql")
	down := migrationSQL(t, "000003_avatar_anniversary_experience.down.sql")
	for _, fragment := range []string{"count_mode", "pinned", "cover_ref", "message", "purpose", "resource_id", "moment_photo"} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{"DROP COLUMN count_mode", "DROP COLUMN pinned", "DROP COLUMN cover_ref", "DROP COLUMN message", "DROP COLUMN purpose", "DROP COLUMN resource_id"} {
		if !strings.Contains(down, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}

func TestPersonalSpacesMigrationShape(t *testing.T) {
	up := migrationSQL(t, "000004_personal_spaces_reflections.up.sql")
	down := migrationSQL(t, "000004_personal_spaces_reflections.down.sql")
	for _, fragment := range []string{
		"space_type", "visibility", "kind", "source_moment_id", "daily_question_id",
		"moment_perspectives", "moment_perspective_photos", "moment_reactions", "daily_questions", "daily_answers",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"DROP TABLE daily_answers", "DROP TABLE daily_questions", "DROP TABLE moment_reactions",
		"DROP TABLE moment_perspective_photos", "DROP TABLE moment_perspectives",
		"DROP COLUMN daily_question_id", "DROP COLUMN source_moment_id", "DROP COLUMN kind",
		"DROP COLUMN visibility", "DROP COLUMN space_type",
	} {
		if !strings.Contains(down, fragment) {
			t.Fatalf("down migration missing %q", fragment)
		}
	}
}

func TestPersonalSpacesMigrationPreservesExistingMoments(t *testing.T) {
	root, err := sql.Open("mysql", migrationTestDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.Ping(); err != nil {
		t.Skipf("migration test MySQL is unavailable: %v", err)
	}

	databaseName := fmt.Sprintf("just_for_you_personal_spaces_%d", time.Now().UnixNano())
	if _, err := root.Exec("CREATE DATABASE " + databaseName + " CHARACTER SET utf8mb4"); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP DATABASE " + databaseName)

	db, err := sql.Open("mysql", migrationDatabaseDSN(databaseName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, name := range []string{
		"000001_init.up.sql", "000002_couple_experience.up.sql", "000003_avatar_anniversary_experience.up.sql",
	} {
		if _, err := db.Exec(migrationSQL(t, name)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}

	occurredAt := time.Date(2026, 10, 20, 9, 30, 0, 0, time.UTC)
	if _, err := db.Exec("INSERT INTO users(id,wechat_openid) VALUES(1,'a'),(2,'b')"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO couples(id,user_a_id,user_b_id,status) VALUES(7,1,2,'active')"); err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec("INSERT INTO moments(couple_id,author_user_id,body,status_code,place_name,occurred_at,created_at,updated_at) VALUES(7,1,'旧记录','','',?,?,?)", occurredAt, occurredAt, occurredAt)
	if err != nil {
		t.Fatal(err)
	}
	momentID, _ := result.LastInsertId()
	if _, err := db.Exec("INSERT INTO moment_photos(moment_id,object_key,sort_order) VALUES(?,?,1)", momentID, "legacy/photo.jpg"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(migrationSQL(t, "000004_personal_spaces_reflections.up.sql")); err != nil {
		t.Fatalf("apply personal spaces migration: %v", err)
	}
	var spaceType, visibility, kind, body, objectKey string
	var authorID, photoCount int64
	var gotOccurredAt time.Time
	err = db.QueryRow(`SELECT m.space_type,m.visibility,m.kind,m.author_user_id,m.body,m.occurred_at,
		(SELECT COUNT(*) FROM moment_photos p WHERE p.moment_id=m.id),
		(SELECT object_key FROM moment_photos p WHERE p.moment_id=m.id LIMIT 1)
		FROM moments m WHERE m.id=?`, momentID).Scan(&spaceType, &visibility, &kind, &authorID, &body, &gotOccurredAt, &photoCount, &objectKey)
	if err != nil {
		t.Fatal(err)
	}
	if spaceType != "shared" || visibility != "couple" || kind != "standard" || authorID != 1 || body != "旧记录" || photoCount != 1 || objectKey != "legacy/photo.jpg" || !gotOccurredAt.Equal(occurredAt) {
		t.Fatalf("migrated moment = space:%s visibility:%s kind:%s author:%d body:%q occurred:%s photos:%d key:%q", spaceType, visibility, kind, authorID, body, gotOccurredAt, photoCount, objectKey)
	}

	if _, err := db.Exec(migrationSQL(t, "000004_personal_spaces_reflections.down.sql")); err != nil {
		t.Fatalf("rollback personal spaces migration: %v", err)
	}
	for _, table := range []string{"moment_perspectives", "moment_perspective_photos", "moment_reactions", "daily_questions", "daily_answers"} {
		assertTableExists(t, db, table, false)
	}
	for _, column := range []string{"space_type", "visibility", "kind", "source_moment_id", "daily_question_id"} {
		assertColumnExists(t, db, "moments", column, false)
	}
}

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
