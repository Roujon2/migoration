package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Init is the command to initialize a new migration project
var InitCmd = &cobra.Command{
	Use:  "init",
	Short: "Initialize a new migoration project",
	Long: `Initialize a new migoration project in the current directory and create a new configuration file for managing database migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Call init migration
		initMigration()
	},
}

// Init function
func init() {
	RootCmd.AddCommand(InitCmd)
}

// initMigration function to initialize
func initMigration() {
	migrationDir := "migrations"
	configFile := "migoration.yaml"

	// Create the migrations directory
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		if err := os.Mkdir(migrationDir, 0755); err != nil {
			log.Fatalf("Error creating migrations directory: %v", err)
			return
		}
		fmt.Println("Created migrations directory.")
	} else {
		fmt.Println("Migrations directory already exists... Continuing")
	}

	// Create the configuration file
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Write the configuration file
		err := os.WriteFile(configFile, []byte(createConfig()), 0644)
		if err != nil {
			log.Fatalf("Error creating config file: %v", err)
			return
		}
		fmt.Println("Created migoration.yaml configuration file.")
	}

	fmt.Println("Migoration initialization complete! Run 'migoration help' to see available commands.")

}


// Function to create the config file info
func createConfig() string {
	config := `# Migoration configuration file
# Modify ${ENV_VAR} placeholders with actual environment variables
database_url: ${DATABASE_URL}
migration_path: migrations`

	return config
}