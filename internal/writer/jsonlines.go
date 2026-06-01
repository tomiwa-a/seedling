package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type JsonLinesWriter struct {
	dir string
}

func NewJsonLinesWriter(dir string) *JsonLinesWriter {
	return &JsonLinesWriter{dir: dir}
}

func (w *JsonLinesWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	filePath := filepath.Join(w.dir, table.Name+".jsonl")
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create jsonl file %s: %w", filePath, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)

	for _, row := range rows {
		obj := make(map[string]any, len(row))
		for k, v := range row {
			obj[k] = v
		}
		if err := encoder.Encode(obj); err != nil {
			return fmt.Errorf("encode json row: %w", err)
		}
	}

	return nil
}

func (w *JsonLinesWriter) Close() error {
	return nil
}

var _ writerinterface.Writer = (*JsonLinesWriter)(nil)
