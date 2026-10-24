package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type migration struct {
	version int
	name    string
}

func ApplyMigrations(
	ctx context.Context,
	db *sql.DB,
	directory, direction string,
	target int,
) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("migration direction must be up or down")
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	files, err := migrations(directory)
	if err != nil {
		return err
	}
	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}
	if direction == "up" {
		return migrateUp(ctx, db, directory, files, applied, target)
	}
	return migrateDown(ctx, db, directory, files, applied, target)
}

func migrations(directory string) ([]migration, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	items := []migration{}
	seen := map[int]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		prefix := strings.SplitN(entry.Name(), "_", 2)[0]
		version, err := strconv.Atoi(prefix)
		if err != nil || version <= 0 {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		if seen[version] {
			return nil, fmt.Errorf("duplicate migration version %d", version)
		}
		seen[version] = true
		items = append(items, migration{version: version, name: entry.Name()})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].version < items[j].version
	})
	return items, nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[int]bool{}
	for rows.Next() {
		var version int
		if err = rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func migrateUp(
	ctx context.Context,
	db *sql.DB,
	directory string,
	files []migration,
	applied map[int]bool,
	target int,
) error {
	for _, item := range files {
		if applied[item.version] || (target > 0 && item.version != target) {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(directory, item.name))
		if err != nil {
			return err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", item.name, err)
		}
		if _, err = tx.ExecContext(
			ctx,
			"INSERT INTO schema_migrations(version,name) VALUES(?,?)",
			item.version,
			item.name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", item.name, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", item.name, err)
		}
	}
	return nil
}

func migrateDown(
	ctx context.Context,
	db *sql.DB,
	directory string,
	files []migration,
	applied map[int]bool,
	target int,
) error {
	sort.Slice(files, func(i, j int) bool {
		return files[i].version > files[j].version
	})
	for _, item := range files {
		if !applied[item.version] || (target > 0 && item.version != target) {
			continue
		}
		name := strings.TrimSuffix(item.name, ".up.sql") + ".down.sql"
		contents, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("rollback migration %s: %w", name, err)
		}
		if _, err = tx.ExecContext(
			ctx,
			"DELETE FROM schema_migrations WHERE version=?",
			item.version,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("remove migration record %d: %w", item.version, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit rollback %s: %w", name, err)
		}
	}
	return nil
}
