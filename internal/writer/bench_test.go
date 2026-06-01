//go:build integration

package writer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomiwa-a/seedling/pkg/schema"
	w "github.com/tomiwa-a/seedling/pkg/writer"
)

func getBenchDSN() string {
	dsn := os.Getenv("SEEDLING_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://pitersonsmartpro@localhost:5432/shard?sslmode=disable"
	}
	return dsn
}

func setupBenchTable(t testing.TB, ctx context.Context, dsn, name string) (*schema.Table, *pgxpool.Pool, func()) {
	t.Helper()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS public."+name)
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE public.%s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			score DOUBLE PRECISION,
			active BOOLEAN
		)
	`, name))
	if err != nil {
		t.Fatal(err)
	}

	tbl := &schema.Table{
		Name: name,
		Columns: []*schema.Column{
			{Name: "name", Type: schema.TypeVarchar},
			{Name: "score", Type: schema.TypeFloat},
			{Name: "active", Type: schema.TypeBoolean},
		},
	}

	cleanup := func() {
		pool.Exec(ctx, "DROP TABLE IF EXISTS public."+name)
		pool.Close()
	}

	return tbl, pool, cleanup
}

func genBenchRows(n int) w.Rows {
	rows := make(w.Rows, n)
	for i := range rows {
		rows[i] = w.Row{
			"name":   fmt.Sprintf("user_%d", i),
			"score":  float64(i) * 1.5,
			"active": i%2 == 0,
		}
	}
	return rows
}

func BenchmarkSqlWriter_10K(b *testing.B) {
	rows := genBenchRows(10000)
	tbl := &schema.Table{
		Name: "bench",
		Columns: []*schema.Column{
			{Name: "name", Type: schema.TypeVarchar},
			{Name: "score", Type: schema.TypeFloat},
			{Name: "active", Type: schema.TypeBoolean},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		sw := NewSqlWriter(&buf, WithBatchSize(1000))
		sw.WriteTable(context.Background(), tbl, rows)
	}
}

func BenchmarkDbWriter_10K(b *testing.B) {
	dsn := getBenchDSN()
	ctx := context.Background()
	tbl, _, cleanup := setupBenchTable(b, ctx, dsn, "bench_db_10k")
	defer cleanup()

	rows := genBenchRows(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dw, _ := NewDbWriter(ctx, dsn, WithDbBatchSize(1000))
		dw.WriteTable(ctx, tbl, rows)
		dw.pool.Exec(ctx, "TRUNCATE public.bench_db_10k")
		dw.Close()
	}
}

func BenchmarkCopyWriter_10K(b *testing.B) {
	dsn := getBenchDSN()
	ctx := context.Background()
	tbl, _, cleanup := setupBenchTable(b, ctx, dsn, "bench_copy_10k")
	defer cleanup()

	rows := genBenchRows(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cw, _ := NewCopyWriter(ctx, dsn)
		cw.WriteTable(ctx, tbl, rows)
		cw.pool.Exec(ctx, "TRUNCATE public.bench_copy_10k")
		cw.Close()
	}
}

func BenchmarkDbWriter_100K(b *testing.B) {
	dsn := getBenchDSN()
	ctx := context.Background()
	tbl, _, cleanup := setupBenchTable(b, ctx, dsn, "bench_db_100k")
	defer cleanup()

	rows := genBenchRows(100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dw, _ := NewDbWriter(ctx, dsn, WithDbBatchSize(1000))
		dw.WriteTable(ctx, tbl, rows)
		dw.pool.Exec(ctx, "TRUNCATE public.bench_db_100k")
		dw.Close()
	}
}

func BenchmarkCopyWriter_100K(b *testing.B) {
	dsn := getBenchDSN()
	ctx := context.Background()
	tbl, _, cleanup := setupBenchTable(b, ctx, dsn, "bench_copy_100k")
	defer cleanup()

	rows := genBenchRows(100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cw, _ := NewCopyWriter(ctx, dsn)
		cw.WriteTable(ctx, tbl, rows)
		cw.pool.Exec(ctx, "TRUNCATE public.bench_copy_100k")
		cw.Close()
	}
}
