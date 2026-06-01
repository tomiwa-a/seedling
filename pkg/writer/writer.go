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

type WriterFunc struct {
	WriteTableFn func(ctx context.Context, table schema.Table, rows Rows) error
	CloseFn      func() error
}

func NewWriterFunc(writeFn func(ctx context.Context, table schema.Table, rows Rows) error) *WriterFunc {
	return &WriterFunc{
		WriteTableFn: writeFn,
		CloseFn:      func() error { return nil },
	}
}

func (f *WriterFunc) WriteTable(ctx context.Context, table schema.Table, rows Rows) error {
	return f.WriteTableFn(ctx, table, rows)
}

func (f *WriterFunc) Close() error {
	return f.CloseFn()
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
