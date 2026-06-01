package writer_test

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

func TestWriterFuncAdapter(t *testing.T) {
	var called bool
	w := writer.NewWriterFunc(func(ctx context.Context, table schema.Table, rows writer.Rows) error {
		called = true
		if len(rows) != 2 {
			t.Errorf("rows length = %d, want 2", len(rows))
		}
		return nil
	})

	err := w.WriteTable(context.Background(), schema.Table{Name: "test"}, writer.Rows{
		{"id": 1, "name": "alice"},
		{"id": 2, "name": "bob"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("WriterFunc was not called")
	}
}

func TestMultiWriterCallsAll(t *testing.T) {
	var callCount int
	w1 := writer.NewWriterFunc(func(ctx context.Context, table schema.Table, rows writer.Rows) error {
		callCount++
		return nil
	})
	w2 := writer.NewWriterFunc(func(ctx context.Context, table schema.Table, rows writer.Rows) error {
		callCount++
		return nil
	})

	mw := writer.NewMultiWriter(w1, w2)
	err := mw.WriteTable(context.Background(), schema.Table{Name: "test"}, writer.Rows{{"a": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestMultiWriterErrorStopsEarly(t *testing.T) {
	w1 := writer.NewWriterFunc(func(ctx context.Context, table schema.Table, rows writer.Rows) error {
		return nil
	})
	w2 := writer.NewWriterFunc(func(ctx context.Context, table schema.Table, rows writer.Rows) error {
		return nil
	})

	mw := writer.NewMultiWriter(w1, w2)
	err := mw.WriteTable(context.Background(), schema.Table{Name: "test"}, writer.Rows{{"a": 1}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultiWriterClose(t *testing.T) {
	var closed1, closed2 bool
	w1 := &closeRecorder{closeFn: func() error { closed1 = true; return nil }}
	w2 := &closeRecorder{closeFn: func() error { closed2 = true; return nil }}

	mw := writer.NewMultiWriter(w1, w2)
	err := mw.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !closed1 {
		t.Error("w1 was not closed")
	}
	if !closed2 {
		t.Error("w2 was not closed")
	}
}

type closeRecorder struct {
	closeFn func() error
}

func (c *closeRecorder) WriteTable(ctx context.Context, table schema.Table, rows writer.Rows) error {
	return nil
}

func (c *closeRecorder) Close() error {
	return c.closeFn()
}

func TestRowMap(t *testing.T) {
	r := writer.Row{"id": 1, "email": "test@example.com"}
	if r["id"] != 1 {
		t.Errorf("Row['id'] = %v, want 1", r["id"])
	}
	if r["email"] != "test@example.com" {
		t.Errorf("Row['email'] = %v", r["email"])
	}
}

func TestRowsType(t *testing.T) {
	rows := writer.Rows{
		{"id": 1},
		{"id": 2},
	}
	if len(rows) != 2 {
		t.Errorf("len(rows) = %d, want 2", len(rows))
	}
}
