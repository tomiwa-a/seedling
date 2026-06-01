package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	internalplanbuilder "github.com/tomiwa-a/seedling/internal/planbuilder"
	internalstream "github.com/tomiwa-a/seedling/internal/stream"
	internalwriter "github.com/tomiwa-a/seedling/internal/writer"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test data",
	Long: `Generate realistic test data based on an introspected
schema and optional generator overrides.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		count, _ := cmd.Flags().GetInt("count")
		output, _ := cmd.Flags().GetString("output")
		schemaFile, _ := cmd.Flags().GetString("schema")
		configFile, _ := cmd.Flags().GetString("config")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		seed, _ := cmd.Flags().GetInt64("seed")

		if seed == 0 {
			seed = time.Now().UnixNano()
		}

		sch, err := loadSchema(schemaFile)
		if err != nil {
			return fmt.Errorf("load schema: %w", err)
		}

		if verbose {
			fmt.Printf("Loaded schema: %s (%d tables)\n", sch.Name, len(sch.Tables))
		}

		pb := internalplanbuilder.New(count)
		plan, err := pb.Build(ctx, sch, nil)
		if err != nil {
			return fmt.Errorf("build plan: %w", err)
		}

		if dryRun {
			fmt.Printf("Plan: %d total rows across %d tables\n", plan.TotalCount, len(plan.Tables))
			for _, tp := range plan.Tables {
				fmt.Printf("  %s: %d rows\n", tp.Table.Name, tp.Count)
			}
			return nil
		}

		sg := internalstream.New()
		sg.SetSeed(uint64(seed))

		if verbose {
			fmt.Printf("Using seed: %d\n", seed)
		}

		for _, tp := range plan.Tables {
			colHints := make(map[string]schema.GeneratorHint)
			for _, col := range tp.Table.Columns {
				if col.Hint != "" && col.Hint != schema.HintAuto {
					colHints[col.Name] = col.Hint
				}
			}
			if len(colHints) > 0 {
				sg.SetHints(tp.Table.Name, colHints)
			}
		}

		if configFile != "" {
			hints, err := loadHints(configFile)
			if err != nil {
				return err
			}
			for _, tp := range plan.Tables {
				if h, ok := hints[tp.Table.Name]; ok {
					for col, hint := range h {
						sg.SetHint(tp.Table.Name, col, hint)
					}
				}
			}
		}

		outFile, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer outFile.Close()

		sw := internalwriter.NewSqlWriter(outFile, internalwriter.WithBatchSize(batchSize))

		if verbose {
			fmt.Printf("Generating %d rows across %d tables...\n", plan.TotalCount, len(plan.Tables))
		}

		if err := sg.Generate(ctx, plan, sw); err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		if verbose {
			fmt.Printf("Output written to %s\n", output)
		}
		return nil
	},
}

type hintsFile map[string]map[string]schema.GeneratorHint

func loadSchema(path string) (*schema.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema file: %w", err)
	}

	var sch schema.Schema
	if err := yaml.Unmarshal(data, &sch); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	return &sch, nil
}

func loadHints(path string) (map[string]map[string]schema.GeneratorHint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var hf hintsFile
	if err := yaml.Unmarshal(data, &hf); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return hf, nil
}

func init() {
	generateCmd.Flags().Int("count", 100, "Number of rows per root table")
	generateCmd.Flags().Int64("seed", 0, "Random seed (0 = random)")
	generateCmd.Flags().String("schema", "schema.yaml", "Schema file path")
	generateCmd.Flags().String("output", "seed.sql", "Output file path")
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
