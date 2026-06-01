package writer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type ParquetWriter struct {
	dir string
}

func NewParquetWriter(dir string) *ParquetWriter {
	return &ParquetWriter{dir: dir}
}

func (w *ParquetWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()
	filePath := filepath.Join(w.dir, table.Name+".parquet")

	var lines []string

	header := strings.Join(cols, "\t")
	lines = append(lines, header)

	for _, row := range rows {
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = parquetFormatValue(row[col])
		}
		lines = append(lines, strings.Join(vals, "\t"))
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write parquet file: %w", err)
	}

	return nil
}

func (w *ParquetWriter) Close() error {
	return nil
}

func parquetFormatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case int32:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case float32:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

var _ writerinterface.Writer = (*ParquetWriter)(nil)
