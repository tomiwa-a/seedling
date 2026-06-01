package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seedling",
	Short: "Relational test data factory",
	Long: `Seedling is a schema-aware test data generator.
It reads your database schema and generates realistic,
referentially-integer test data at scale.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
