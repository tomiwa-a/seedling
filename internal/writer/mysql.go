package writer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

	maxLengths := make(map[string]int)
	numericBounds := make(map[string][2]float64)
	for _, col := range table.Columns {
		if col.MaxLength > 0 {
			maxLengths[col.Name] = col.MaxLength
		}
		if col.NumericPrec > 0 {
			scale := col.NumericScale
			prec := col.NumericScale
			if prec == 0 {
				prec = col.NumericPrec
			}
			maxVal := 1.0
			for i := 0; i < prec; i++ {
				maxVal *= 10
			}
			if scale > 0 {
				maxVal = maxVal - 1
				scaleFactor := 1.0
				for i := 0; i < scale; i++ {
					scaleFactor *= 10
				}
				maxVal = maxVal / scaleFactor
			} else {
				maxVal = maxVal - 1
			}
			minVal := -maxVal
			numericBounds[col.Name] = [2]float64{minVal, maxVal}
		}
	}

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
				val := row[col]
				if ml, ok := maxLengths[col]; ok {
					val = truncateStringValue(val, ml)
				}
				if bounds, ok := numericBounds[col]; ok {
					val = clampNumericValue(val, bounds[0], bounds[1])
				}
				writeMySQLValue(&buf, val)
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
	_, err := w.db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0")
	if err != nil {
		return fmt.Errorf("disable FK checks: %w", err)
	}
	defer w.db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1")

	_, err = w.db.ExecContext(ctx, "TRUNCATE TABLE "+table.Name)
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
		if isISODateTime(val) {
			writeMySQLString(buf, strings.ReplaceAll(strings.TrimSuffix(val, "Z"), "T", " "))
		} else {
			writeMySQLString(buf, val)
		}
	case time.Time:
		writeMySQLString(buf, val.Format("2006-01-02 15:04:05"))
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

func isISODateTime(s string) bool {
	return len(s) >= 19 && s[4] == '-' && s[7] == '-' && s[10] == 'T' && s[13] == ':'
}

func truncateStringValue(v any, maxLen int) any {
	if v == nil {
		return v
	}
	if s, ok := v.(string); ok && len(s) > maxLen {
		return s[:maxLen]
	}
	return v
}

func clampNumericValue(v any, min, max float64) any {
	if v == nil {
		return v
	}
	var f float64
	switch val := v.(type) {
	case int:
		f = float64(val)
	case int64:
		f = float64(val)
	case int32:
		f = float64(val)
	case float64:
		f = val
	case float32:
		f = float64(val)
	default:
		return v
	}
	if f < min {
		return min
	}
	if f > max {
		return max
	}
	return f
}
