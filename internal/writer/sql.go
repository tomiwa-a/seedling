package writer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/tomiwa-a/seedling/pkg/schema"
	writerinterface "github.com/tomiwa-a/seedling/pkg/writer"
)

type SqlWriter struct {
	w       io.Writer
	buf     *bytes.Buffer
	schema  string
	batchSz int
}

type SqlOption func(*SqlWriter)

func WithSchema(s string) SqlOption {
	return func(w *SqlWriter) { w.schema = s }
}

func WithBatchSize(n int) SqlOption {
	return func(w *SqlWriter) { w.batchSz = n }
}

func NewSqlWriter(w io.Writer, opts ...SqlOption) *SqlWriter {
	sw := &SqlWriter{
		w:       w,
		buf:     &bytes.Buffer{},
		batchSz: 50,
	}
	for _, opt := range opts {
		opt(sw)
	}
	return sw
}

func (s *SqlWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	if len(rows) == 0 {
		return nil
	}

	cols := table.ColumnNames()

	tableName := table.Name
	if s.schema != "" {
		tableName = s.schema + "." + tableName
	}

	for i := 0; i < len(rows); i += s.batchSz {
		end := i + s.batchSz
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		var buf bytes.Buffer
		buf.WriteString("INSERT INTO ")
		buf.WriteString(tableName)
		buf.WriteString(" (")
		for j, c := range cols {
			if j > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c)
		}
		buf.WriteString(") VALUES\n")

		for j, row := range batch {
			if j > 0 {
				buf.WriteString(",\n")
			}
			buf.WriteByte('(')
			for k, col := range cols {
				if k > 0 {
					buf.WriteString(", ")
				}
				s.writeValue(&buf, row[col])
			}
			buf.WriteByte(')')
		}
		buf.WriteString(";\n\n")

		if _, err := s.w.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("write insert: %w", err)
		}
	}

	return nil
}

func (s *SqlWriter) Close() error {
	return nil
}

func (s *SqlWriter) writeValue(buf *bytes.Buffer, v any) {
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
		s.writeString(buf, val)
	case time.Time:
		s.writeString(buf, val.Format("2006-01-02 15:04:05"))
	case []byte:
		fmt.Fprintf(buf, `E'\\x%x'`, val)
	default:
		s.writeString(buf, fmt.Sprintf("%v", val))
	}
}

func (s *SqlWriter) writeString(buf *bytes.Buffer, str string) {
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

func (s *SqlWriter) String() string {
	return s.buf.String()
}

var _ writerinterface.Writer = (*SqlWriter)(nil)

type TrimWriter struct {
	inner writerinterface.Writer
	cols  map[string][]string
}

func NewTrimWriter(inner writerinterface.Writer) *TrimWriter {
	return &TrimWriter{inner: inner, cols: make(map[string][]string)}
}

func (tw *TrimWriter) Trim(table string, columns []string) {
	tw.cols[table] = columns
}

func (tw *TrimWriter) WriteTable(ctx context.Context, table *schema.Table, rows writerinterface.Rows) error {
	trimmed := make(writerinterface.Rows, len(rows))
	drop := tw.cols[table.Name]

	for i, row := range rows {
		r := make(writerinterface.Row)
		for k, v := range row {
			skip := false
			for _, d := range drop {
				if k == d {
					skip = true
					break
				}
			}
			if !skip {
				r[k] = v
			}
		}
		trimmed[i] = r
	}

	tc := *table
	tc.Columns = nil
	for _, col := range table.Columns {
		drop := false
		for _, d := range tw.cols[table.Name] {
			if col.Name == d {
				drop = true
				break
			}
		}
		if !drop {
			tc.Columns = append(tc.Columns, col)
		}
	}

	return tw.inner.WriteTable(ctx, &tc, trimmed)
}

func (tw *TrimWriter) Close() error {
	return tw.inner.Close()
}
