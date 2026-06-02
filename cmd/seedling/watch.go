package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch schema and regenerate on change",
	Long: `Monitor the schema file and generator config for
changes and automatically regenerate test data.

Uses fsnotify to watch files. Changes are debounced
(default 2s) to avoid rapid regeneration during saves.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaFile, _ := cmd.Flags().GetString("schema")
		generatorsFile, _ := cmd.Flags().GetString("generators")
		output, _ := cmd.Flags().GetString("output")
		dbDSN, _ := cmd.Flags().GetString("db")
		count, _ := cmd.Flags().GetInt("count")
		seed, _ := cmd.Flags().GetInt64("seed")
		formatStr, _ := cmd.Flags().GetString("format")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		useCopy, _ := cmd.Flags().GetBool("copy")
		truncate, _ := cmd.Flags().GetBool("truncate")
		parallel, _ := cmd.Flags().GetBool("parallel")
		verbose, _ := cmd.Flags().GetBool("verbose")
		debounce, _ := cmd.Flags().GetDuration("debounce")

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("create watcher: %w", err)
		}
		defer watcher.Close()

		if err := watcher.Add(schemaFile); err != nil {
			return fmt.Errorf("watch schema file: %w", err)
		}
		if generatorsFile != "" {
			if err := watcher.Add(generatorsFile); err != nil {
				return fmt.Errorf("watch generators file: %w", err)
			}
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		var debounceTimer *time.Timer

		if verbose {
			fmt.Printf("Watching %s", schemaFile)
			if generatorsFile != "" {
				fmt.Printf(", %s", generatorsFile)
			}
			fmt.Printf(" (debounce: %s)...\n", debounce)
		}

		runOnce := func() {
			if verbose {
				fmt.Println("\nRegenerating...")
			}
			if err := runGenerate(generateParams{
				ctx:            ctx,
				schemaFile:     schemaFile,
				generatorsFile: generatorsFile,
				output:         output,
				formatStr:      formatStr,
				count:          count,
				seed:           seed,
				dbDSN:          dbDSN,
				useCopy:        useCopy,
				truncate:       truncate,
				parallel:       parallel,
				dryRun:         false,
				verbose:        verbose,
				batchSize:      batchSize,
			}); err != nil {
				log.Printf("Generate error: %v", err)
			}
		}

		runOnce()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(debounce, runOnce)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				log.Printf("Watch error: %v", err)

			case <-sigCh:
				fmt.Println("\nShutting down...")
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				return nil
			}
		}
	},
}

func init() {
	watchCmd.Flags().String("schema", "schema.yaml", "Schema file path to watch")
	watchCmd.Flags().String("generators", "", "Path to YAML generator config file")
	watchCmd.Flags().String("output", "seed.sql", "Output file path")
	watchCmd.Flags().String("db", "", "Database DSN for direct insert")
	watchCmd.Flags().Int("count", 100, "Number of rows per root table")
	watchCmd.Flags().Int64("seed", 0, "Random seed (0 = random)")
	watchCmd.Flags().String("format", "sql", "Output format (sql, csv, jsonl, parquet)")
	watchCmd.Flags().Int("batch-size", 1000, "Rows per batch")
	watchCmd.Flags().Bool("copy", false, "Use COPY protocol (Postgres only)")
	watchCmd.Flags().Bool("truncate", false, "TRUNCATE tables before inserting")
	watchCmd.Flags().Bool("parallel", false, "Generate independent tables in parallel (breaks determinism)")
	watchCmd.Flags().Bool("verbose", false, "Detailed progress output")
	watchCmd.Flags().Duration("debounce", 2*time.Second, "Debounce interval between regenerations")
	rootCmd.AddCommand(watchCmd)
}
