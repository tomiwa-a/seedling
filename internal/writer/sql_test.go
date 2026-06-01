package writer

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

func TestSqlWriterSingleRow(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "email", Type: schema.TypeVarchar},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1), "email": "alice@test.com"},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}
	err = sw.Close()
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "INSERT INTO") {
		t.Error("missing INSERT")
	}
	if !strings.Contains(out, "users") {
		t.Error("missing table name")
	}
	if !strings.Contains(out, "'alice@test.com'") {
		t.Error("missing email value")
	}
}

func TestSqlWriterMultiRow(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "items",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "label", Type: schema.TypeVarchar},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1), "label": "a"},
		{"id": int64(2), "label": "b"},
		{"id": int64(3), "label": "c"},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "VALUES\n(1, 'a'),\n(2, 'b'),\n(3, 'c')") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestSqlWriterEscaping(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "t",
		Columns: []schema.Column{
			{Name: "v", Type: schema.TypeText},
		},
	}

	rows := writerinterface.Rows{
		{"v": "it's a test"},
		{"v": "line1\nline2"},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "'it''s a test'") {
		t.Errorf("expected escaped quote, got: %s", out)
	}
}

func TestSqlWriterTypes(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "types",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeBigSerial},
			{Name: "active", Type: schema.TypeBoolean},
			{Name: "score", Type: schema.TypeFloat},
			{Name: "created", Type: schema.TypeTimestamp},
		},
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	rows := writerinterface.Rows{
		{"id": int64(1), "active": true, "score": float64(3.14), "created": now},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "TRUE") {
		t.Error("missing TRUE")
	}
	if !strings.Contains(out, "3.14") {
		t.Error("missing float")
	}
	if !strings.Contains(out, "'2024-01-15 10:30:00'") {
		t.Error("missing timestamp")
	}
}

func TestSqlWriterNullable(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "t",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "optional", Type: schema.TypeVarchar, Nullable: true},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1), "optional": nil},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "NULL") {
		t.Error("missing NULL")
	}
}

func TestSqlWriterBatchSize(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf, WithBatchSize(2))

	tbl := schema.Table{
		Name: "t",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1)},
		{"id": int64(2)},
		{"id": int64(3)},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Count(out, "INSERT INTO") != 2 {
		t.Errorf("expected 2 INSERT statements, got %d", strings.Count(out, "INSERT INTO"))
	}
}

func TestSqlWriterSchemaPrefix(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf, WithSchema("public"))

	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1)},
	}

	err := sw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "public.users") {
		t.Errorf("expected public.users, got: %s", out)
	}
}

func TestSqlWriterEmptyRows(t *testing.T) {
	var buf bytes.Buffer
	sw := NewSqlWriter(&buf)

	tbl := schema.Table{
		Name: "empty",
	}

	err := sw.WriteTable(context.Background(), tbl, writerinterface.Rows{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() > 0 {
		t.Error("expected no output for empty rows")
	}
}

func TestTrimWriter(t *testing.T) {
	var buf bytes.Buffer
	inner := NewSqlWriter(&buf)
	tw := NewTrimWriter(inner)
	tw.Trim("users", []string{"secret"})

	tbl := schema.Table{
		Name: "users",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeSerial},
			{Name: "secret", Type: schema.TypeVarchar},
			{Name: "email", Type: schema.TypeVarchar},
		},
	}

	rows := writerinterface.Rows{
		{"id": int64(1), "secret": "hidden", "email": "a@b.com"},
	}

	err := tw.WriteTable(context.Background(), tbl, rows)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Error("secret column should be trimmed")
	}
	if !strings.Contains(out, "email") {
		t.Error("email column should be present")
	}
	if !strings.Contains(out, "'a@b.com'") {
		t.Error("email value should be present")
	}
}
