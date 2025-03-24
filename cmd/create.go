package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var migrationType string


// CreateCmd - Command to create a new database migration
var CreateCmd = &cobra.Command{
	Use: "create [name]",
	Short: "Create a new database migration",
	Long: `Create a new database migration with the given name. The migration will be created in the migrations directory specified in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("Error: Migration name must be specified with the -m flag.")
			return
		}
		// Parse spaces in name into '_'
		name = strings.ReplaceAll(name, " ", "_")
		
		// Call create migration
		createMigration(name, migrationType)
	},
}

// Init function
func init() {
	RootCmd.AddCommand(CreateCmd)
	// Save the flag in the migrationType variable
	CreateCmd.Flags().StringVarP(&migrationType, "type", "t", "sql", "Type of migration file to create (sql or go)")

	// Define -m flag for migration name
	CreateCmd.Flags().StringP("name", "m", "", "Name of the migration")
}

// Create migration
func createMigration(name string, migrationType string) {
	// Extract config and check if migrations dir exists
	config, err := loadConfig("migoration")
	if err != nil {
		fmt.Printf("Error loading configuration file 'migoration.yaml': %v\n", err)
		return
	}
	migrations_dir := config.MigrationsDir
	if _, err := os.Stat(migrations_dir); os.IsNotExist(err) {
		fmt.Printf("Error: Migrations directory '%s' does not exist in current directory\n", migrations_dir)
		return
	}

	// Version timestamp
	timestamp := time.Now().UTC().Format("20060102150405")

	// Create the migration file based on migration type
	switch migrationType {
	case "sql":
		createSQLMigration(name, timestamp, migrations_dir)
	case "go":
		createSQLMigration(name, timestamp, migrations_dir) // TODO
	default:
		fmt.Printf("Error: Invalid migration type '%s'. Use 'sql' or 'go'\n", migrationType)
	}
}


// Function to create a SQL migration file
func createSQLMigration(name string, version string, migrations_dir string) {
	// Create the SQL migration files
	upFile := filepath.Join(migrations_dir, fmt.Sprintf("%s_%s_up.sql", version, name))
	downFile := filepath.Join(migrations_dir, fmt.Sprintf("%s_%s_down.sql", version, name))

	// Up migration template
	upTemplate := fmt.Sprintf(`-- Migration Up: %s_%s
-- Write your UP migration SQL here
`, version, name)

	// Down migration template
	downTemplate := fmt.Sprintf(`-- Migration Down: %s_%s
-- Write your DOWN migration SQL here (rollback)
`, version, name)

	// Write the files
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		fmt.Printf("Error creating up migration file: %v\n", err)
		return
	}
	if err :=
		os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		fmt.Printf("Error creating down migration file: %v\n", err)
		return
	}

	fmt.Printf("Migration files created: %s, %s\n", upFile, downFile)
}






