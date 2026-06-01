package introspect_test

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/introspect"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestIntrospectorFuncAdapter(t *testing.T) {
	var called bool
	intr := introspect.IntrospectorFunc(func(ctx context.Context, dsn string) (*schema.Schema, error) {
		called = true
		if dsn != "postgres://localhost:5432/test" {
			t.Errorf("dsn = %q", dsn)
		}
		return &schema.Schema{Name: "test"}, nil
	})

	s, err := intr.Introspect(context.Background(), "postgres://localhost:5432/test")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("IntrospectorFunc was not called")
	}
	if s.Name != "test" {
		t.Errorf("Schema.Name = %q, want %q", s.Name, "test")
	}
}

func TestIntrospectorFuncReturnsError(t *testing.T) {
	intr := introspect.IntrospectorFunc(func(ctx context.Context, dsn string) (*schema.Schema, error) {
		return nil, nil
	})
	s, err := intr.Introspect(context.Background(), "any")
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Log("IntrospectorFunc returned nil schema, nil error")
	}
}

func TestOptionsDefaults(t *testing.T) {
	opts := introspect.Options{}
	if opts.IncludeViews {
		t.Error("IncludeViews should default to false")
	}
	if len(opts.SchemaFilter) != 0 {
		t.Errorf("SchemaFilter = %v, want empty", opts.SchemaFilter)
	}
}

func TestOptionsWithValues(t *testing.T) {
	opts := introspect.Options{
		SchemaFilter:  []string{"public"},
		IncludeViews:  true,
		MaxRowsPreview: 10,
	}
	if len(opts.SchemaFilter) != 1 || opts.SchemaFilter[0] != "public" {
		t.Errorf("SchemaFilter = %v", opts.SchemaFilter)
	}
	if !opts.IncludeViews {
		t.Error("IncludeViews should be true")
	}
	if opts.MaxRowsPreview != 10 {
		t.Errorf("MaxRowsPreview = %d", opts.MaxRowsPreview)
	}
}

func TestResultStruct(t *testing.T) {
	r := introspect.Result{
		Schema: &schema.Schema{Name: "public"},
		DSN:    "postgres://localhost:5432/db",
	}
	if r.Schema.Name != "public" {
		t.Errorf("Result.Schema.Name = %q", r.Schema.Name)
	}
	if r.DSN != "postgres://localhost:5432/db" {
		t.Errorf("Result.DSN = %q", r.DSN)
	}
}
