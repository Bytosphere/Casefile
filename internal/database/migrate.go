package database

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations
var migrationFS embed.FS

// Migration represents a single SQL file.
type Migration struct {
	Version int
	Name    string
	Path    string
}

// Apply applies the migration by loading its source.
func (m *Migration) Apply(db *Database) error {
	source, err := migrationFS.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(source))
	if err != nil {
		return err
	}

	return nil
}

// Migrate performs database migrations from a local `migrations/` file.
func Migrate(db *Database) error {
	// Create the migration table.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS __Migration (
			version    INTEGER PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %v", err)
	}

	for _, migration := range migrations {
		tx, err := db.BeginTransaction()
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		migrationErr := migration.Apply(db)
		if migrationErr != nil {
			return fmt.Errorf("migration '%s': %v", migration.Name, migrationErr)
		}

		// For each successful migration, add it to the history table.
		_, txErr := tx.Exec(
			"INSERT INTO __Migration (version, name, applied_at) VALUES (?, ?, ?)",
			migration.Version,
			migration.Name,
			time.Now().UTC().Format(time.RFC3339),
		)
		if txErr != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("migration '%s': rollback failed: %v", migration.Name, rollbackErr)
			}
			return fmt.Errorf("migration '%s': insert failed: %v", migration.Name, txErr)
		}

		if txErr = tx.Commit(); txErr != nil {
			return fmt.Errorf("migration '%s': commit failed: %v", migration.Name, txErr)
		}

		fmt.Printf("applied migration: V%d '%s'\n", migration.Version, migration.Name)
	}

	return nil
}

// loadMigrations loads all migration files from the embedded migrations directory.
// It validates the files against the migration history.
func loadMigrations(db *Database) ([]Migration, error) {
	files, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration

	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse version from filename (e.g., V1_Init.sql -> version 1).
		baseName := strings.TrimSuffix(name, ".sql")
		if !strings.HasPrefix(baseName, "V") {
			continue
		}

		// Remove 'V'.
		baseName = baseName[1:]

		underscoreIdx := strings.Index(baseName, "_")
		if underscoreIdx == -1 {
			continue
		}

		versionStr := baseName[:underscoreIdx]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version from %s: %w", name, err)
		}

		migrationName := baseName[underscoreIdx+1:]

		migrations = append(migrations, Migration{
			Version: version,
			Name:    migrationName,
			Path:    "migrations/" + name,
		})
	}

	// Query already applied migrations.
	rows, err := db.connection.Query("SELECT version FROM __Migration")
	if err != nil {
		return nil, fmt.Errorf("failed to query migration history: %w", err)
	}
	defer rows.Close()

	appliedVersions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		appliedVersions[version] = true
	}

	// Filter out already applied migrations.
	var pendingMigrations []Migration
	for _, m := range migrations {
		if !appliedVersions[m.Version] {
			pendingMigrations = append(pendingMigrations, m)
		}
	}

	// Sort by version ascending.
	sort.Slice(pendingMigrations, func(i, j int) bool {
		return pendingMigrations[i].Version < pendingMigrations[j].Version
	})

	return pendingMigrations, nil
}
