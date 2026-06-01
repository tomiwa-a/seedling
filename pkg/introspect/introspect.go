package introspect

import (
	"context"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

type Result struct {
	Schema *schema.Schema
	DSN    string
}

type Introspector interface {
	Introspect(ctx context.Context, dsn string) (*schema.Schema, error)
}

type IntrospectorFunc func(ctx context.Context, dsn string) (*schema.Schema, error)

func (f IntrospectorFunc) Introspect(ctx context.Context, dsn string) (*schema.Schema, error) {
	return f(ctx, dsn)
}

type Options struct {
	SchemaFilter   []string
	TableFilter    []string
	ExcludeTables  []string
	IncludeViews   bool
	MaxRowsPreview int
}
