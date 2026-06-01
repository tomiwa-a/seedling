package writer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type dbBase struct {
	pool    *pgxpool.Pool
	schema  string
	batchSz int
}

type DbWriter struct {
	dbBase
}

type CopyWriter struct {
	dbBase
}

type DbOption func(*dbBase)

func WithDbSchema(s string) DbOption {
	return func(b *dbBase) { b.schema = s }
}

func WithDbBatchSize(n int) DbOption {
	return func(b *dbBase) { b.batchSz = n }
}

func NewDbWriter(ctx context.Context, dsn string, opts ...DbOption) (*DbWriter, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	b := &dbBase{pool: pool, batchSz: 1000}
	for _, opt := range opts {
		opt(b)
	}

	return &DbWriter{dbBase: *b}, nil
}

func (w *DbWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()
	tableName := table.Name
	if w.schema != "" {
		tableName = w.schema + "." + tableName
	}

	for i := 0; i < len(rows); i += w.batchSz {
		end := i + w.batchSz
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		var buf strings.Builder
		buf.WriteString("INSERT INTO ")
		buf.WriteString(tableName)
		buf.WriteString(" (")
		for j, c := range cols {
			if j > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c)
		}
		buf.WriteString(") VALUES ")

		for j, row := range batch {
			if j > 0 {
				buf.WriteString(", ")
			}
			buf.WriteByte('(')
			for k, col := range cols {
				if k > 0 {
					buf.WriteString(", ")
				}
				writeSQLValue(&buf, row[col])
			}
			buf.WriteByte(')')
		}

		if _, err := w.pool.Exec(ctx, buf.String()); err != nil {
			return fmt.Errorf("insert into %s: %w", tableName, err)
		}
	}

	return nil
}

func (w *DbWriter) Close() error {
	w.pool.Close()
	return nil
}

func (w *DbWriter) Truncate(ctx context.Context, table *schema.Table) error {
	tableName := table.Name
	if w.schema != "" {
		tableName = w.schema + "." + tableName
	}
	_, err := w.pool.Exec(ctx, "TRUNCATE TABLE "+tableName+" CASCADE")
	return err
}

func (w *DbWriter) Pool() *pgxpool.Pool {
	return w.pool
}

func NewCopyWriter(ctx context.Context, dsn string, opts ...DbOption) (*CopyWriter, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	b := &dbBase{pool: pool}
	for _, opt := range opts {
		opt(b)
	}

	return &CopyWriter{dbBase: *b}, nil
}

func (w *CopyWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()
	tableName := table.Name
	if w.schema != "" {
		tableName = w.schema + "." + tableName
	}

	copyRows := &copyRowSource{cols: cols, rows: rows}
	_, err := w.pool.CopyFrom(ctx,
		pgx.Identifier{tableName},
		cols,
		copyRows,
	)
	if err != nil {
		return fmt.Errorf("copy into %s: %w", tableName, err)
	}

	return nil
}

func (w *CopyWriter) Close() error {
	w.pool.Close()
	return nil
}

func (w *CopyWriter) Truncate(ctx context.Context, table *schema.Table) error {
	tableName := table.Name
	if w.schema != "" {
		tableName = w.schema + "." + tableName
	}
	_, err := w.pool.Exec(ctx, "TRUNCATE TABLE "+tableName+" CASCADE")
	return err
}

func (w *CopyWriter) Pool() *pgxpool.Pool {
	return w.pool
}

type copyRowSource struct {
	cols []string
	rows writerinterface.Rows
	pos  int
}

func (s *copyRowSource) Next() bool {
	return s.pos < len(s.rows)
}

func (s *copyRowSource) Values() ([]any, error) {
	row := s.rows[s.pos]
	s.pos++

	vals := make([]any, len(s.cols))
	for i, col := range s.cols {
		vals[i] = row[col]
	}
	return vals, nil
}

func (s *copyRowSource) Err() error {
	return nil
}

var _ writerinterface.Writer = (*DbWriter)(nil)
var _ writerinterface.Writer = (*CopyWriter)(nil)
var _ pgx.CopyFromSource = (*copyRowSource)(nil)

func writeSQLValue(buf *strings.Builder, v any) {
	if v == nil {
		buf.WriteString("NULL")
		return
	}

	switch val := v.(type) {
	case int:
		fmt.Fprintf(buf, "%d", val)
	case int64:
		fmt.Fprintf(buf, "%d", val)
	case int32:
		fmt.Fprintf(buf, "%d", val)
	case float64:
		fmt.Fprintf(buf, "%g", val)
	case float32:
		fmt.Fprintf(buf, "%g", val)
	case bool:
		if val {
			buf.WriteString("TRUE")
		} else {
			buf.WriteString("FALSE")
		}
	case string:
		writeSQLString(buf, val)
	case time.Time:
		writeSQLString(buf, val.Format("2006-01-02 15:04:05"))
	case []byte:
		fmt.Fprintf(buf, `E'\\x%x'`, val)
	default:
		writeSQLString(buf, fmt.Sprintf("%v", val))
	}
}

func writeSQLString(buf *strings.Builder, str string) {
	buf.WriteByte('\'')
	for _, c := range str {
		if c == '\'' {
			buf.WriteString("''")
		} else if c == '\\' {
			buf.WriteString("\\\\")
		} else {
			buf.WriteRune(c)
		}
	}
	buf.WriteByte('\'')
}
