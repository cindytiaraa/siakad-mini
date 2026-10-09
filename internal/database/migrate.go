package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(100) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`)
	return err
}

func appliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// runFile menjalankan semua statement di file SQL + pencatatan versi dalam satu transaction.
func runFile(db *sql.DB, path, bookkeeping, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range strings.Split(string(content), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(bookkeeping, version); err != nil {
		return err
	}
	return tx.Commit()
}

// MigrateUp menjalankan semua migration *.up.sql yang belum pernah dijalankan.
func MigrateUp(db *sql.DB, dir string) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}
	for _, f := range files {
		version := strings.TrimSuffix(filepath.Base(f), ".up.sql")
		if applied[version] {
			continue
		}
		if err := runFile(db, f, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("migration %s gagal: %w", version, err)
		}
		fmt.Println("migrated up  :", version)
	}
	return nil
}

// MigrateDown membatalkan migration terakhir (atau semuanya jika all = true).
func MigrateDown(db *sql.DB, dir string, all bool) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}
	for {
		var version string
		err := db.QueryRow(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		file := filepath.Join(dir, version+".down.sql")
		if err := runFile(db, file, `DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
			return fmt.Errorf("rollback %s gagal: %w", version, err)
		}
		fmt.Println("migrated down:", version)
		if !all {
			return nil
		}
	}
}
