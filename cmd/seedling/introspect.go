package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	internalintrospect "github.com/tomiwa-a/seedling/internal/introspect"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect",
	Short: "Introspect a database schema",
	Long: `Connect to a database, read its schema (tables, columns,
foreign keys, constraints), and output the schema definition.
Supports PostgreSQL and MySQL/MariaDB.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, _ := cmd.Flags().GetString("db")
		output, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if format == "" {
			ext := strings.ToLower(filepath.Ext(output))
			switch ext {
			case ".json":
				format = "json"
			default:
				format = "yaml"
			}
		}

		ctx := cmd.Context()

		var schema *schema.Schema
		var err error

		if isMysqlDSN(db) {
			introspector, err := internalintrospect.NewMysqlIntrospector(ctx, db)
			if err != nil {
				return fmt.Errorf("create mysql introspector: %w", err)
			}
			defer introspector.Close()
			schema, err = introspector.Introspect(ctx, db)
		} else {
			introspector, err := internalintrospect.NewPostgresIntrospector(ctx, db)
			if err != nil {
				return fmt.Errorf("create postgres introspector: %w", err)
			}
			defer introspector.Close()
			schema, err = introspector.Introspect(ctx, db)
		}

		if err != nil {
			return fmt.Errorf("introspect: %w", err)
		}

		var data []byte
		switch format {
		case "json":
			data, err = json.MarshalIndent(schema, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
		default:
			data, err = yaml.Marshal(schema)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
		}

		if err := os.WriteFile(output, data, 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		fmt.Printf("Schema written to %s (%s)\n", output, format)
		return nil
	},
}

func isMysqlDSN(dsn string) bool {
	lower := strings.ToLower(dsn)
	return strings.Contains(lower, "tcp(") ||
		strings.Contains(lower, "mysql") ||
		strings.Contains(lower, "mariadb") ||
		(!strings.Contains(lower, "postgres://") &&
			!strings.Contains(lower, "postgresql://") &&
			strings.Contains(lower, "@"))
}

func init() {
	introspectCmd.Flags().String("db", "", "Database DSN (required)")
	introspectCmd.Flags().String("output", "schema.yaml", "Output file path")
	introspectCmd.Flags().String("format", "", "Output format (yaml, json)")
	introspectCmd.MarkFlagRequired("db")
	rootCmd.AddCommand(introspectCmd)
}
