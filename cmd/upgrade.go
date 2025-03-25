package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Roujon2/migoration/db"
	"github.com/jackc/pgx/v5"

	"github.com/spf13/cobra"
)

// Upgrade command
var UpgradeCmd = &cobra.Command{
	Use:   "upgrade [target]",
	Short: "Run database migrations to a specific target",
	Long:  `Run database migrations to a specific target version. The target version is the number of steps or 'head' to migrate to.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		runUpgrade(target)
	},
}

func init() {
	RootCmd.AddCommand(UpgradeCmd)
}

// Run the migoration upgrade
func runUpgrade(target string) {
	// Check if it's a positive number
	if _, err := strconv.Atoi(target); err != nil && target != "head" {
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
		// If there are no rows, set the current version to 0
		if pgx.ErrNoRows == err {
			currentVersion = "0"
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
	currentIndex := 0
	for i, migration := range migrations {
		if migration.Version == currentVersion {
			currentIndex = i + 1
			break
		}
	}

	// Slice the migrations to apply
	var migrationsToApply Migrations
	if target == "head" {
		migrationsToApply = migrations[currentIndex:]
	} else {
		// Parse the target to int
		target, err := strconv.Atoi(target)
		if err != nil {
			fmt.Printf("Target must be a number or 'head'\n")
			return
		}
		// Cap the target to the number of migrations
		if currentIndex+target > len(migrations) {
			target = len(migrations) - currentIndex
		}
		migrationsToApply = migrations[currentIndex : currentIndex+target]
	}

	// Apply the migrations
	var prev *Migration
	if currentIndex > 0 {
		prev = &migrations[currentIndex-1]
	}
	for i := range migrationsToApply {
		migration := &migrationsToApply[i]
		err := applyMigration(db, migration, "up", prev)
		if err != nil {
			fmt.Printf("Error applying migration '%s': %v\n", migration.Version, err)
			return
		}
		prev = migration
	}

	fmt.Printf("Migrations applied successfully\n")
}
