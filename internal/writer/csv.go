package writer

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type CsvWriter struct {
	dir       string
	delimiter rune
	quoting   bool
}

type CsvOption func(*CsvWriter)

func WithCsvDelimiter(d rune) CsvOption {
	return func(w *CsvWriter) { w.delimiter = d }
}

func WithCsvQuoting(q bool) CsvOption {
	return func(w *CsvWriter) { w.quoting = q }
}

func NewCsvWriter(dir string, opts ...CsvOption) *CsvWriter {
	w := &CsvWriter{
		dir:       dir,
		delimiter: ',',
		quoting:   true,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *CsvWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()

	filePath := filepath.Join(w.dir, table.Name+".csv")
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create csv file %s: %w", filePath, err)
	}
	defer f.Close()

	encoder := csv.NewWriter(f)
	encoder.Comma = w.delimiter

	if err := encoder.Write(cols); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := make([]string, len(cols))
		for i, col := range cols {
			record[i] = csvFormatValue(row[col])
		}
		if err := encoder.Write(record); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	encoder.Flush()
	return encoder.Error()
}

func (w *CsvWriter) Close() error {
	return nil
}

func csvFormatValue(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case bool:
		return strconv.FormatBool(val)
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

var _ writerinterface.Writer = (*CsvWriter)(nil)
