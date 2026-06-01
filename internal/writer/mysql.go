package writer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type MysqlWriter struct {
	db      *sql.DB
	batchSz int
}

type MysqlOption func(*MysqlWriter)

func WithMysqlBatchSize(n int) MysqlOption {
	return func(w *MysqlWriter) { w.batchSz = n }
}

func NewMysqlWriter(ctx context.Context, dsn string, opts ...MysqlOption) (*MysqlWriter, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	w := &MysqlWriter{
		db:      db,
		batchSz: 1000,
	}
	for _, opt := range opts {
		opt(w)
	}

	return w, nil
}

func (w *MysqlWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()

	for i := 0; i < len(rows); i += w.batchSz {
		end := i + w.batchSz
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		var buf strings.Builder
		buf.WriteString("INSERT INTO ")
		buf.WriteString(table.Name)
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
				writeMySQLValue(&buf, row[col])
			}
			buf.WriteByte(')')
		}

		if _, err := w.db.ExecContext(ctx, buf.String()); err != nil {
			return fmt.Errorf("insert into %s: %w", table.Name, err)
		}
	}

	return nil
}

func (w *MysqlWriter) Close() error {
	return w.db.Close()
}

func (w *MysqlWriter) Truncate(ctx context.Context, table *schema.Table) error {
	_, err := w.db.ExecContext(ctx, "TRUNCATE TABLE "+table.Name)
	return err
}

func (w *MysqlWriter) DB() *sql.DB {
	return w.db
}

var _ writerinterface.Writer = (*MysqlWriter)(nil)

func writeMySQLValue(buf *strings.Builder, v any) {
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
			buf.WriteString("1")
		} else {
			buf.WriteString("0")
		}
	case string:
		writeMySQLString(buf, val)
	case []byte:
		buf.WriteString("X'")
		for _, b := range val {
			fmt.Fprintf(buf, "%02X", b)
		}
		buf.WriteByte('\'')
	default:
		writeMySQLString(buf, fmt.Sprintf("%v", val))
	}
}

func writeMySQLString(buf *strings.Builder, str string) {
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
