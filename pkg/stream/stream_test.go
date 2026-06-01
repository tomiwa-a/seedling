package stream_test

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/stream"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

func TestStreamGeneratorFuncAdapter(t *testing.T) {
	var called bool
	sg := stream.StreamGeneratorFunc(func(ctx context.Context, p *plan.Plan, w writer.Writer) error {
		called = true
		if p.TotalCount != 100 {
			t.Errorf("TotalCount = %d", p.TotalCount)
		}
		return nil
	})

	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{Name: "test"}, Count: 100},
		},
		TotalCount: 100,
	}

	w := writer.NewWriterFunc(func(ctx context.Context, table *schema.Table, rows writer.Rows) error {
		return nil
	})

	err := sg.Generate(context.Background(), p, w)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("StreamGeneratorFunc was not called")
	}
}

func TestProgressEventTypes(t *testing.T) {
	events := []stream.ProgressEvent{
		{Type: stream.EventTableStart, TableName: "users"},
		{Type: stream.EventTableProgress, TableName: "users", RowsSoFar: 50, TotalRows: 100},
		{Type: stream.EventTableComplete, TableName: "users", RowsSoFar: 100, TotalRows: 100},
		{Type: stream.EventError, Error: nil},
		{Type: stream.EventComplete},
	}

	if events[0].Type != stream.EventTableStart {
		t.Errorf("expected EventTableStart")
	}
	if events[1].RowsSoFar != 50 {
		t.Errorf("RowsSoFar = %d", events[1].RowsSoFar)
	}
	if events[2].RowsSoFar != 100 {
		t.Errorf("RowsSoFar = %d", events[2].RowsSoFar)
	}
	if events[4].Type != stream.EventComplete {
		t.Errorf("expected EventComplete")
	}
}

func TestOptionsStruct(t *testing.T) {
	eventCh := make(chan stream.ProgressEvent, 10)
	opts := stream.Options{
		Concurrency: 4,
		Events:      eventCh,
	}
	if opts.Concurrency != 4 {
		t.Errorf("Concurrency = %d", opts.Concurrency)
	}
	if opts.Events == nil {
		t.Error("Events channel should not be nil")
	}
	close(eventCh)
}

func TestEventTypeValues(t *testing.T) {
	if int(stream.EventTableStart) != 0 {
		t.Errorf("EventTableStart = %d, want 0", stream.EventTableStart)
	}
	if int(stream.EventComplete) != 4 {
		t.Errorf("EventComplete = %d, want 4", stream.EventComplete)
	}
}
