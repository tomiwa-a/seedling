package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch schema and regenerate on change",
	Long: `Monitor the database schema and generator files for
changes and automatically regenerate test data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, _ := cmd.Flags().GetString("db")
		generators, _ := cmd.Flags().GetString("generators")
		output, _ := cmd.Flags().GetString("output")
		fmt.Printf("watch: db=%s generators=%s output=%s (not yet implemented)\n", db, generators, output)
		return nil
	},
}

func init() {
	watchCmd.Flags().String("db", "", "Database DSN (required)")
	watchCmd.Flags().String("generators", "", "Path to Go generator files")
	watchCmd.Flags().String("output", "seed.sql", "Output file")
	watchCmd.MarkFlagRequired("db")
	rootCmd.AddCommand(watchCmd)
}
