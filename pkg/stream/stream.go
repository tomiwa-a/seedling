package stream

import (
	"context"

	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

type EventType int

const (
	EventTableStart EventType = iota
	EventTableProgress
	EventTableComplete
	EventError
	EventComplete
)

type ProgressEvent struct {
	Type       EventType `json:"type"`
	TableName  string    `json:"table_name,omitempty"`
	RowsSoFar  int64     `json:"rows_so_far,omitempty"`
	TotalRows  int64     `json:"total_rows,omitempty"`
	Error      error     `json:"error,omitempty"`
}

type StreamGenerator interface {
	Generate(ctx context.Context, p *plan.Plan, w writer.Writer) error
}

type StreamGeneratorFunc func(ctx context.Context, p *plan.Plan, w writer.Writer) error

func (f StreamGeneratorFunc) Generate(ctx context.Context, p *plan.Plan, w writer.Writer) error {
	return f(ctx, p, w)
}

type Options struct {
	Concurrency int
	Events      chan<- ProgressEvent
}
