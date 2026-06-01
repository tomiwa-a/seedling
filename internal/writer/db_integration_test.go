//go:build integration

package writer

import (
	"context"
	"os"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
	w "github.com/tomiwa-a/seedling/pkg/writer"
)

func getTestDSN() string {
	dsn := os.Getenv("SEEDLING_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://pitersonsmartpro@localhost:5432/shard?sslmode=disable"
	}
	return dsn
}

func TestDbWriterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn := getTestDSN()

	dw, err := NewDbWriter(ctx, dsn, WithDbSchema("public"))
	if err != nil {
		t.Fatal(err)
	}
	defer dw.Close()

	_, err = dw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test")
	if err != nil {
		t.Fatal(err)
	}

	_, err = dw.pool.Exec(ctx, `
		CREATE TABLE public.seedling_test (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			active BOOLEAN
		)
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer dw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test")

	rows := w.Rows{
		{"name": "Alice", "active": true},
		{"name": "Bob", "active": false},
		{"name": "Charlie", "active": true},
	}

	insertCols := &schema.Table{
		Name: "seedling_test",
		Columns: []*schema.Column{
			{Name: "name", Type: schema.TypeVarchar},
			{Name: "active", Type: schema.TypeBoolean},
		},
	}

	err = dw.WriteTable(ctx, insertCols, rows)
	if err != nil {
		t.Fatal(err)
	}

	var count int
	err = dw.pool.QueryRow(ctx, "SELECT COUNT(*) FROM public.seedling_test").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

func TestCopyWriterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn := getTestDSN()

	cw, err := NewCopyWriter(ctx, dsn, WithDbSchema("public"))
	if err != nil {
		t.Fatal(err)
	}
	defer cw.Close()

	_, err = cw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test_copy")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cw.pool.Exec(ctx, `
		CREATE TABLE public.seedling_test_copy (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			score DOUBLE PRECISION
		)
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer cw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test_copy")

	rows := w.Rows{
		{"name": "Item A", "score": 3.14},
		{"name": "Item B", "score": 2.71},
		{"name": "Item C", "score": 1.41},
		{"name": "Item D", "score": 9.81},
		{"name": "Item E", "score": 6.67},
	}

	insertCols := &schema.Table{
		Name: "seedling_test_copy",
		Columns: []*schema.Column{
			{Name: "name", Type: schema.TypeVarchar},
			{Name: "score", Type: schema.TypeFloat},
		},
	}

	err = cw.WriteTable(ctx, insertCols, rows)
	if err != nil {
		t.Fatal(err)
	}

	var count int
	err = cw.pool.QueryRow(ctx, "SELECT COUNT(*) FROM public.seedling_test_copy").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 5 {
		t.Errorf("expected 5 rows, got %d", count)
	}
}

func TestTruncateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn := getTestDSN()

	dw, err := NewDbWriter(ctx, dsn, WithDbSchema("public"))
	if err != nil {
		t.Fatal(err)
	}
	defer dw.Close()

	tbl := &schema.Table{
		Name: "seedling_test_truncate",
		Columns: []*schema.Column{
			{Name: "id", Type: schema.TypeSerial},
		},
		Constraints: []*schema.Constraint{
			{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
		},
	}

	_, err = dw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test_truncate")
	if err != nil {
		t.Fatal(err)
	}

	_, err = dw.pool.Exec(ctx, `
		CREATE TABLE public.seedling_test_truncate (
			id SERIAL PRIMARY KEY
		)
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer dw.pool.Exec(ctx, "DROP TABLE IF EXISTS public.seedling_test_truncate")

	insertCols := &schema.Table{
		Name: "seedling_test_truncate",
		Columns: []*schema.Column{
			{Name: "id", Type: schema.TypeInteger},
		},
	}

	for i := 0; i < 5; i++ {
		dw.WriteTable(ctx, insertCols, w.Rows{{"id": i + 1}})
	}

	var count int
	dw.pool.QueryRow(ctx, "SELECT COUNT(*) FROM public.seedling_test_truncate").Scan(&count)
	if count != 5 {
		t.Fatalf("expected 5 rows before truncate, got %d", count)
	}

	err = dw.Truncate(ctx, tbl)
	if err != nil {
		t.Fatal(err)
	}

	dw.pool.QueryRow(ctx, "SELECT COUNT(*) FROM public.seedling_test_truncate").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 rows after truncate, got %d", count)
	}
}
