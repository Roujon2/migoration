package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Base command for the CLI
var RootCmd = &cobra.Command{
	Use:   "migoration",
	Short: "A simple migration tool in Golang",
	Long:  `Migoration is a simple migration tool for PostgreSQL databases in Golang applications`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Migoration CLI: Use 'migoration --help' for available commands")
	},
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
