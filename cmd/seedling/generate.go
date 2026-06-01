package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	internalplanbuilder "github.com/tomiwa-a/seedling/internal/planbuilder"
	internalstream "github.com/tomiwa-a/seedling/internal/stream"
	internalwriter "github.com/tomiwa-a/seedling/internal/writer"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
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
		dbDSN, _ := cmd.Flags().GetString("db")
		useCopy, _ := cmd.Flags().GetBool("copy")
		truncate, _ := cmd.Flags().GetBool("truncate")
		parallel, _ := cmd.Flags().GetBool("parallel")

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

		pb := internalplanbuilder.NewWithSeed(count, seed)
		plan, err := pb.Build(ctx, sch, nil)
		if err != nil {
			return fmt.Errorf("build plan: %w", err)
		}

		if dryRun {
			fmt.Printf("Plan: %d total rows across %d tables\n", plan.TotalCount, len(plan.Tables))
			for _, tp := range plan.Tables {
				fmt.Printf("  %s: %d rows\n", tp.Table.Name, tp.Count)
			}
			if plan.CircularGroup != nil {
				fmt.Printf("Circular FKs detected: %d pass1 tables, %d pass2 tables\n",
					len(plan.CircularGroup.Pass1Tables), len(plan.CircularGroup.Pass2Tables))
			}
			return nil
		}

		sg := internalstream.New()
		sg.SetSeed(uint64(seed))
		sg.SetParallel(parallel)

		if verbose {
			fmt.Printf("Using seed: %d\n", seed)
			sg.SetProgress(func(p internalstream.Progress) {
				pct := float64(p.RowsWritten) / float64(p.TotalRows) * 100
				fmt.Printf("\r  [%s] %d/%d rows (%.0f%%) | %.0f rows/sec | ETA %s",
					p.Table, p.RowsWritten, p.TotalRows, pct,
					p.RowsPerSec, p.ETA.Truncate(time.Second))
				if p.RowsWritten >= p.TotalRows {
					fmt.Println()
				}
			})
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

		var w writerinterface.Writer

		if dbDSN != "" {
			if verbose {
				fmt.Printf("Connecting to: %s\n", dbDSN)
			}

			var dbWriter writerinterface.Writer
			if useCopy {
				dbWriter, err = internalwriter.NewCopyWriter(ctx, dbDSN,
					internalwriter.WithDbSchema(sch.Name),
					internalwriter.WithDbBatchSize(batchSize))
			} else {
				dbWriter, err = internalwriter.NewDbWriter(ctx, dbDSN,
					internalwriter.WithDbSchema(sch.Name),
					internalwriter.WithDbBatchSize(batchSize))
			}
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer dbWriter.Close()

			if truncate {
				if verbose {
					fmt.Println("Truncating tables...")
				}
				for _, tp := range plan.Tables {
					if truncatable, ok := dbWriter.(interface {
						Truncate(context.Context, *schema.Table) error
					}); ok {
						if err := truncatable.Truncate(ctx, tp.Table); err != nil {
							return fmt.Errorf("truncate %s: %w", tp.Table.Name, err)
						}
					}
				}
			}

			w = dbWriter
		} else {
			format, _ := cmd.Flags().GetString("format")

			switch format {
			case "csv":
				if err := os.MkdirAll(output, 0755); err != nil {
					return fmt.Errorf("create output dir: %w", err)
				}
				w = internalwriter.NewCsvWriter(output)
			case "jsonl", "jsonlines":
				if err := os.MkdirAll(output, 0755); err != nil {
					return fmt.Errorf("create output dir: %w", err)
				}
				w = internalwriter.NewJsonLinesWriter(output)
			case "parquet":
				if err := os.MkdirAll(output, 0755); err != nil {
					return fmt.Errorf("create output dir: %w", err)
				}
				w = internalwriter.NewParquetWriter(output)
			default:
				outFile, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("create output: %w", err)
				}
				defer outFile.Close()
				w = internalwriter.NewSqlWriter(outFile, internalwriter.WithBatchSize(batchSize))
			}
		}

		if verbose {
			fmt.Printf("Generating %d rows across %d tables...\n", plan.TotalCount, len(plan.Tables))
		}

		start := time.Now()
		if err := sg.Generate(ctx, plan, w); err != nil {
			return fmt.Errorf("generate: %w", err)
		}
		elapsed := time.Since(start)

		if verbose {
			if dbDSN != "" {
				fmt.Printf("Inserted %d rows in %s (%.0f rows/sec)\n",
					plan.TotalCount, elapsed.Truncate(time.Millisecond),
					float64(plan.TotalCount)/elapsed.Seconds())
			} else {
				fmt.Printf("Output written to %s (%s)\n", output, elapsed.Truncate(time.Millisecond))
			}
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
	generateCmd.Flags().String("db", "", "Database DSN for direct insert")
	generateCmd.Flags().Bool("copy", false, "Use COPY protocol (Postgres only)")
	generateCmd.Flags().Bool("truncate", false, "TRUNCATE tables before inserting")
	generateCmd.Flags().String("generators", "", "Path to Go generator files")
	generateCmd.Flags().String("config", "", "Path to seedling.yaml config")
	generateCmd.Flags().String("preset", "", "Use a saved preset")
	generateCmd.Flags().Bool("dry-run", false, "Print plan without generating")
	generateCmd.Flags().Bool("parallel", false, "Generate independent tables in parallel (breaks determinism)")
	generateCmd.Flags().Bool("verbose", false, "Detailed progress output")
	rootCmd.AddCommand(generateCmd)
}
