package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test data",
	Long: `Generate realistic test data based on an introspected
schema and optional generator overrides.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")
		seed, _ := cmd.Flags().GetInt64("seed")
		db, _ := cmd.Flags().GetString("db")
		output, _ := cmd.Flags().GetString("output")
		fmt.Printf("generate: count=%d seed=%d db=%s output=%s (not yet implemented)\n", count, seed, db, output)
		return nil
	},
}

func init() {
	generateCmd.Flags().Int("count", 100, "Number of rows per root table")
	generateCmd.Flags().Int64("seed", 0, "Random seed (0 = random)")
	generateCmd.Flags().String("db", "", "Database DSN for direct insert")
	generateCmd.Flags().String("output", "seed.sql", "Output file or directory")
	generateCmd.Flags().String("format", "sql", "Output format (sql, csv, jsonl, parquet)")
	generateCmd.Flags().Int("batch-size", 1000, "Rows per batch")
	generateCmd.Flags().Bool("copy", false, "Use COPY protocol (Postgres only)")
	generateCmd.Flags().Bool("truncate", false, "TRUNCATE tables before inserting")
	generateCmd.Flags().String("generators", "", "Path to Go generator files")
	generateCmd.Flags().String("config", "", "Path to seedling.yaml config")
	generateCmd.Flags().String("preset", "", "Use a saved preset")
	generateCmd.Flags().Bool("dry-run", false, "Print plan without generating")
	generateCmd.Flags().Bool("verbose", false, "Detailed progress output")
	rootCmd.AddCommand(generateCmd)
}
