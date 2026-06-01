package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect",
	Short: "Introspect a database schema",
	Long: `Connect to a database, read its schema (tables, columns,
foreign keys, constraints), and output the schema definition.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, _ := cmd.Flags().GetString("db")
		output, _ := cmd.Flags().GetString("output")
		fmt.Printf("introspect: db=%s output=%s (not yet implemented)\n", db, output)
		return nil
	},
}

func init() {
	introspectCmd.Flags().String("db", "", "Database DSN (required)")
	introspectCmd.Flags().String("output", "schema.yaml", "Output file path")
	introspectCmd.MarkFlagRequired("db")
	rootCmd.AddCommand(introspectCmd)
}
