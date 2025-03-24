package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Struct representing a database migration
type Migration struct{
	Version string `json:"version"`
	Name string `json:"name"`
	UpFile string `json:"up_file"`
	DownFile string `json:"down_file"`
	IsApplied bool `json:"is_applied"`
	AppliedAt time.Time `json:"applied_at"`
}

// Set up the migrations version table in the database
func setupVersionTable(conn *pgx.Conn) error {
	// Create the migrations table
	_, err := conn.Exec(context.Background(),`
		CREATE TABLE IF NOT EXISTS migoration_version (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);`)
	
	return err
}


// Apply a single migration
func applyMigration(conn *pgx.Conn, migration *Migration, direction string) error {
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

	// Update the version table by removing any existing row and adding the current one
	if _, err := conn.Exec(ctx, "DELETE FROM migoration_version WHERE version = $1", migration.Version); err != nil {
		return fmt.Errorf("error deleting old migration version '%s': %v", migration.Version, err)
	}
	if _, err := conn.Exec(ctx, "INSERT INTO migoration_version (version, name) VALUES ($1, $2)", migration.Version, migration.Name); err != nil {
		return fmt.Errorf("error updating migration version '%s': %v", migration.Version, err)
	}

	return nil
}


// Retrieve all migrations from the migrations directory in order
func getMigrations(migrationsDir string) ([]*Migration, error) {
	// Get all files in the migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory '%s': %v", migrationsDir, err)
	}

	// Create an empty slice of migrations
	migrations := []*Migration{}

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
		migration := &Migration{
			Version: parts[0],
			Name: strings.Join(parts[1:len(parts)-1], "_"),
			UpFile: migrationsDir + "/" + fileName,
			DownFile: migrationsDir + "/" + strings.Join(parts[0:len(parts)-1], "_") + "_down.sql",
		}

		// Append to the migrations slice
		migrations = append(migrations, migration)
	}

	// Sort the migrations by version (oldest to newest)
	sort.Slice(migrations, func(i, j int) bool {
		// Define the expected timestamp layout.
		const layout = "20060102150405"
		t1, err1 := time.Parse(layout, migrations[i].Version)
		t2, err2 := time.Parse(layout, migrations[j].Version)
		// Fallback to string comparison if parsing fails.
		if err1 != nil || err2 != nil {
			return migrations[i].Version < migrations[j].Version
		}
		return t1.Before(t2)
	})

	return migrations, nil
}
