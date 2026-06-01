package writer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
	w "github.com/tomiwa-a/seedling/pkg/writer"
)

func TestParquetWriterBasic(t *testing.T) {
	dir := t.TempDir()
	pw := NewParquetWriter(dir)

	tbl := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "name", Type: schema.TypeVarchar},
		},
	}

	rows := w.Rows{
		{"id": int64(1), "name": "Alice"},
		{"id": int64(2), "name": "Bob"},
	}

	err := pw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "users.parquet"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "id\tname") {
		t.Error("missing header")
	}
	if !strings.Contains(content, "1\tAlice") {
		t.Error("missing row data")
	}
}

func TestParquetWriterEmpty(t *testing.T) {
	dir := t.TempDir()
	pw := NewParquetWriter(dir)

	tbl := &schema.Table{
		Name: "empty",
	}

	err := pw.WriteTable(context.Background(), tbl, w.Rows{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "empty.parquet")); !os.IsNotExist(err) {
		t.Error("should not create file for empty rows")
	}
}
