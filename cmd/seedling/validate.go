package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate generator configuration against schema",
	Long: `Check that generator configurations are compatible with
the database schema. Detects stale column references, type
mismatches, and missing required generators.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, _ := cmd.Flags().GetString("db")
		generators, _ := cmd.Flags().GetString("generators")
		fmt.Printf("validate: db=%s generators=%s (not yet implemented)\n", db, generators)
		return nil
	},
}

func init() {
	validateCmd.Flags().String("db", "", "Database DSN")
	validateCmd.Flags().String("generators", "", "Path to Go generator files")
	rootCmd.AddCommand(validateCmd)
}
