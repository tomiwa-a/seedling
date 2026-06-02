package main

import (
	internalstream "github.com/tomiwa-a/seedling/internal/stream"

	"github.com/cheggaaa/pb/v3"
)

func newProgressCallback(verbose bool) internalstream.ProgressFunc {
	var bar *pb.ProgressBar
	var currentTable string

	return func(p internalstream.Progress) {
		if !verbose {
			return
		}
		if p.Table != currentTable {
			if bar != nil {
				bar.Finish()
			}
			currentTable = p.Table
			bar = pb.New64(p.TotalRows)
			bar.SetTemplateString(`{{string . "table"}} {{bar . }} {{percent . }} {{counters . }} {{speed . }}`)
			bar.Set("table", truncateTableName(p.Table, 20))
			bar.Start()
		}
		if bar != nil {
			bar.SetCurrent(p.RowsWritten)
			if p.RowsWritten >= p.TotalRows {
				bar.Finish()
				bar = nil
			}
		}
	}
}

func truncateTableName(name string, max int) string {
	if len(name) <= max {
		return name
	}
	return name[:max-3] + "..."
}
