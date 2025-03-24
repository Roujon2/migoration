package cmd

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/Roujon2/migoration/db"
	"github.com/jackc/pgx/v5"

	"github.com/spf13/cobra"
)

// Upgrade command
var DowngradeCmd = &cobra.Command{
	Use:   "downgrade [target]",
	Short: "Run database downgrade migrations to a specific target",
	Long:  `Run database downgrade migrations to a specific target version. The target version is the number of steps or 'base' to downgrade to.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		runDowngrade(target)
	},
}

func init() {
	RootCmd.AddCommand(DowngradeCmd)
}

// Run the migoration upgrade
func runDowngrade(target string) {
	// Check if it's a positive number
	if _, err := strconv.Atoi(target); err != nil && target != "base" {
		fmt.Printf("Error: Target must be a positive number or 'head'\n")
		return
	}

	// Extract config and check if migrations dir exists
	config, err := loadConfig("migoration")
	if err != nil {
		fmt.Printf("Error loading configuration file 'migoration.yaml': %v\n", err)
		return
	}
	migrations_dir := config.MigrationsDir

	// Connect to the database
	db, err := db.ConnectDB(config.DatabaseURL)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}
	defer db.Close(context.Background())

	// Check if version table exists
	err = setupVersionTable(db)
	if err != nil {
		fmt.Printf("Error setting up version table: %v\n", err)
		return
	}

	// Retrieve current version
	var currentVersion string
	err = db.QueryRow(context.Background(), "SELECT version FROM migoration_version ORDER BY version DESC LIMIT 1").Scan(&currentVersion)
	if err != nil {
		// If there are no rows, assume we're at base
		if pgx.ErrNoRows == err {
			fmt.Printf("No migrations to downgrade: migoration_version table is empty\n")
			return
		} else {
			fmt.Printf("Error retrieving current version: %v\n", err)
			return
		}
	}

	// Retrieve all migrations
	migrations, err := getMigrations(migrations_dir)
	if err != nil {
		fmt.Printf("Error retrieving migrations: %v\n", err)
		return
	}

	// Find the current migration index
	currentIndex := -1
	for i, migration := range migrations {
		if migration.Version == currentVersion {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		fmt.Printf("Error: Current version '%s' not found in migrations\n", currentVersion)
		return
	}

	// Slice the migrations to apply
	var migrationsToApply Migrations
	if target == "base" {
		migrationsToApply = migrations[:currentIndex+1]
		// Flip the migrations to apply
		sort.Sort(sort.Reverse(migrationsToApply))
	} else {
		// Parse the target to int
		target, err := strconv.Atoi(target)
		if err != nil {
			fmt.Printf("Target must be a number or 'base'\n")
			return
		}
		// Cap the target to the base
		if target > currentIndex {
			target = currentIndex
		}
		migrationsToApply = migrations[currentIndex-target : currentIndex+1]
		// Flip the migrations to apply
		sort.Sort(sort.Reverse(migrationsToApply))
	}

	// Apply the migrations
	for i, migration := range migrationsToApply {

		// Handle previous migration for tracking downgrade
		var previousMigration *Migration
		if i > 0 {
			previousMigration = &migrationsToApply[i-1]
		} else {
			previousMigration = nil
		}

		err := applyMigration(db, &migration, "down", previousMigration)
		if err != nil {
			fmt.Printf("Error applying migration '%s': %v\n", migration.Version, err)
			return
		}
	}

	fmt.Printf("Migrations applied successfully\n")
}
