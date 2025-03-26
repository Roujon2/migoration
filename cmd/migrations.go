package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Struct representing a database migration
type Migration struct {
	Version  string `json:"version"`
	Name     string `json:"name"`
	UpFile   string `json:"up_file"`
	DownFile string `json:"down_file"`
}

// ------------------------------ IMPLEMENT SORT INTERFACE ------------------------------
type Migrations []Migration

// Len
func (m Migrations) Len() int {
	return len(m)
}

// Less
func (m Migrations) Less(i, j int) bool {
	// Convert version from string to int
	v1, err := strconv.Atoi(m[i].Version)
	if err != nil {
		return false
	}
	v2, err := strconv.Atoi(m[j].Version)
	if err != nil {
		return false
	}

	return v1 < v2
}

// Swap
func (m Migrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Deep copy of migrations
func (m Migrations) Copy() Migrations {
	copied := make(Migrations, len(m))
	copy(copied, m)
	return copied
}

// Set up the migrations version table in the database
func setupVersionTable(conn *pgx.Conn) error {
	// Create the migrations table
	_, err := conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS migoration_version (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);`)

	return err
}

// Apply a single migration
func applyMigration(conn *pgx.Conn, migration *Migration, direction string, previousVersion *Migration) error {
	ctx := context.Background()

	// Determine file and direction to use
	var filePath string
	var action string

	if direction == "up" {
		filePath = migration.UpFile
		action = "Applying"
	} else {
		filePath = migration.DownFile
		action = "Rolling back"
	}

	fmt.Printf("%s migration %s, version %s\n", action, migration.Name, migration.Version)

	if !strings.HasSuffix(filePath, ".sql") {
		return fmt.Errorf("invalid migration file type '%s': only .sql files are supported", filePath)
	}

	// Read the SQL file
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading migration file '%s': %v", filePath, err)
	}

	// Execute the SQL
	if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("error executing migration file '%s': %v", filePath, err)
	}

	if direction == "up" {
		// If previous version is nil, assume we're at base
		if previousVersion != nil {
			// Update the version table by removing any existing row and adding the current one
			if _, err := conn.Exec(ctx, "DELETE FROM migoration_version WHERE version = $1", previousVersion.Version); err != nil {
				return fmt.Errorf("error deleting old migration version '%s': %v", migration.Version, err)
			}
		}
		if _, err := conn.Exec(ctx, "INSERT INTO migoration_version (version, name) VALUES ($1, $2)", migration.Version, migration.Name); err != nil {
			return fmt.Errorf("error updating migration version '%s': %v", migration.Version, err)
		}
	} else if direction == "down" {
		// Update the version table by removing the current row and adding the previous one
		if _, err := conn.Exec(ctx, "DELETE FROM migoration_version WHERE version = $1", migration.Version); err != nil {
			return fmt.Errorf("error deleting old migration version '%s': %v", migration.Version, err)
		}
		// If previous version is nil, assume we're at base
		if previousVersion == nil {
			return nil
		}
		if _, err := conn.Exec(ctx, "INSERT INTO migoration_version (version, name) VALUES ($1, $2)", previousVersion.Version, previousVersion.Name); err != nil {
			return fmt.Errorf("error updating migration version '%s': %v", previousVersion.Version, err)
		}
	}

	return nil
}

// Retrieve all migrations from the migrations directory in order
func getMigrations(migrationsDir string) (Migrations, error) {
	// Get all files in the migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory '%s': %v", migrationsDir, err)
	}

	// Create an empty slice of migrations
	migrations := Migrations{}

	// Loop through the files
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Parse the file name
		fileName := file.Name()
		parts := strings.Split(fileName, "_")
		if len(parts) < 3 {
			fmt.Printf("Skipping invalid migration file '%s'\n", fileName)
			continue
		}

		// If it's a down file, skip
		if parts[len(parts)-1] == "down.sql" {
			continue
		}

		// Create a new migration
		migration := Migration{
			Version:  parts[0],
			Name:     strings.Join(parts[1:len(parts)-1], "_"),
			UpFile:   migrationsDir + "/" + fileName,
			DownFile: migrationsDir + "/" + strings.Join(parts[0:len(parts)-1], "_") + "_down.sql",
		}

		// Append to the migrations slice
		migrations = append(migrations, migration)
	}

	// Sort the migrations
	sort.Sort(migrations)

	return migrations, nil
}
