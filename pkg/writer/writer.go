package writer

import (
	"context"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

type Row map[string]any

type Rows []Row

type TableResult struct {
	Table schema.Table
	Rows  Rows
}

type Writer interface {
	WriteTable(ctx context.Context, table schema.Table, rows Rows) error
	Close() error
}

type WriterFunc func(ctx context.Context, table schema.Table, rows Rows) error

func (f WriterFunc) WriteTable(ctx context.Context, table schema.Table, rows Rows) error {
	return f(ctx, table, rows)
}

type MultiWriter struct {
	writers []Writer
}

func NewMultiWriter(writers ...Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) WriteTable(ctx context.Context, table schema.Table, rows Rows) error {
	for _, w := range mw.writers {
		if err := w.WriteTable(ctx, table, rows); err != nil {
			return err
		}
	}
	return nil
}

func (mw *MultiWriter) Close() error {
	for _, w := range mw.writers {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}
