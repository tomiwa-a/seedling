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

func TestCsvWriterBasic(t *testing.T) {
	dir := t.TempDir()
	cw := NewCsvWriter(dir)

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

	err := cw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "users.csv"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "id,name") {
		t.Error("missing header")
	}
	if !strings.Contains(content, "1,Alice") {
		t.Error("missing row data")
	}
}

func TestCsvWriterEmpty(t *testing.T) {
	dir := t.TempDir()
	cw := NewCsvWriter(dir)

	tbl := &schema.Table{
		Name: "empty",
	}

	err := cw.WriteTable(context.Background(), tbl, w.Rows{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "empty.csv")); !os.IsNotExist(err) {
		t.Error("should not create file for empty rows")
	}
}

func TestCsvWriterCustomDelimiter(t *testing.T) {
	dir := t.TempDir()
	cw := NewCsvWriter(dir, WithCsvDelimiter('\t'))

	tbl := &schema.Table{
		Name: "t",
		Columns: []*schema.Column{
			{Name: "a", Type: schema.TypeVarchar},
			{Name: "b", Type: schema.TypeVarchar},
		},
	}

	rows := w.Rows{{"a": "hello", "b": "world"}}
	cw.WriteTable(context.Background(), tbl, rows)

	data, _ := os.ReadFile(filepath.Join(dir, "t.csv"))
	content := string(data)
	t.Logf("CSV output: %q", content)
	if !strings.Contains(content, "\t") {
		t.Error("expected tab delimiter in output")
	}
	if !strings.Contains(content, "hello") {
		t.Error("missing data")
	}
}
