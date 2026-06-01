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

func TestJsonLinesWriterBasic(t *testing.T) {
	dir := t.TempDir()
	jw := NewJsonLinesWriter(dir)

	tbl := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "name", Type: schema.TypeVarchar},
			{Name: "active", Type: schema.TypeBoolean},
		},
	}

	rows := w.Rows{
		{"id": int64(1), "name": "Alice", "active": true},
		{"id": int64(2), "name": "Bob", "active": false},
	}

	err := jw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "users.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], `"name":"Alice"`) {
		t.Error("missing name field")
	}
	if !strings.Contains(lines[0], `"active":true`) {
		t.Error("missing active field")
	}
}

func TestJsonLinesWriterEmpty(t *testing.T) {
	dir := t.TempDir()
	jw := NewJsonLinesWriter(dir)

	tbl := &schema.Table{
		Name: "empty",
	}

	err := jw.WriteTable(context.Background(), tbl, w.Rows{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "empty.jsonl")); !os.IsNotExist(err) {
		t.Error("should not create file for empty rows")
	}
}

func TestJsonLinesWriterSpecialChars(t *testing.T) {
	dir := t.TempDir()
	jw := NewJsonLinesWriter(dir)

	tbl := &schema.Table{
		Name: "t",
		Columns: []*schema.Column{
			{Name: "v", Type: schema.TypeText},
		},
	}

	rows := w.Rows{{"v": `hello "world" with 'quotes'`}}
	jw.WriteTable(context.Background(), tbl, rows)

	data, _ := os.ReadFile(filepath.Join(dir, "t.jsonl"))
	content := string(data)
	if !strings.Contains(content, `hello \"world\" with 'quotes'`) {
		t.Errorf("expected JSON-escaped quotes, got: %s", content)
	}
}
