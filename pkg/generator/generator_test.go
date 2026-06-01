package generator_test

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/generator"
)

type testRowContext struct {
	values map[string]any
	index  int64
	table  string
}

func (t testRowContext) Column(name string) (any, bool) {
	v, ok := t.values[name]
	return v, ok
}

func (t testRowContext) RowIndex() int64  { return t.index }
func (t testRowContext) TableName() string { return t.table }

func TestGeneratorFuncAdapter(t *testing.T) {
	var called bool
	g := generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		called = true
		return "hello", nil
	})

	row := testRowContext{values: map[string]any{}, index: 0, table: "test"}
	val, err := g.Generate(context.Background(), row)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("GeneratorFunc was not called")
	}
	if val != "hello" {
		t.Errorf("Generate() = %v, want %v", val, "hello")
	}
}

func TestGeneratorFuncPassesContext(t *testing.T) {
	g := generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		name, ok := row.Column("name")
		if !ok {
			return "unknown", nil
		}
		return name, nil
	})

	row := testRowContext{values: map[string]any{"name": "alice"}, index: 0, table: "test"}
	val, err := g.Generate(context.Background(), row)
	if err != nil {
		t.Fatal(err)
	}
	if val != "alice" {
		t.Errorf("Generate() = %v, want %v", val, "alice")
	}
}

func TestGeneratorFuncMissingColumn(t *testing.T) {
	g := generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		v, ok := row.Column("missing")
		if ok {
			t.Error("expected Column() to return false for missing key")
		}
		if v != nil {
			t.Errorf("expected nil value, got %v", v)
		}
		return "fallback", nil
	})

	row := testRowContext{values: map[string]any{}, index: 0, table: "test"}
	val, err := g.Generate(context.Background(), row)
	if err != nil {
		t.Fatal(err)
	}
	if val != "fallback" {
		t.Errorf("Generate() = %v, want %v", val, "fallback")
	}
}

func TestRowContextReturnsRowIndex(t *testing.T) {
	row := testRowContext{values: map[string]any{}, index: 42, table: "test"}
	if row.RowIndex() != 42 {
		t.Errorf("RowIndex() = %d, want 42", row.RowIndex())
	}
}

func TestRowContextReturnsTableName(t *testing.T) {
	row := testRowContext{values: map[string]any{}, index: 0, table: "orders"}
	if row.TableName() != "orders" {
		t.Errorf("TableName() = %q, want %q", row.TableName(), "orders")
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := generator.NewRegistry()
	g := generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		return "val", nil
	})

	reg.Register("test_gen", g)
	got, ok := reg.Get("test_gen")
	if !ok {
		t.Fatal("expected generator to be found")
	}
	val, err := got.Generate(context.Background(), testRowContext{})
	if err != nil {
		t.Fatal(err)
	}
	if val != "val" {
		t.Errorf("Generate() = %v, want %v", val, "val")
	}
}

func TestRegistryGetMissing(t *testing.T) {
	reg := generator.NewRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("expected false for missing generator")
	}
}

func TestRegistryNames(t *testing.T) {
	reg := generator.NewRegistry()
	reg.Register("a", generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		return nil, nil
	}))
	reg.Register("b", generator.GeneratorFunc(func(ctx context.Context, row generator.RowContext) (any, error) {
		return nil, nil
	}))

	names := reg.Names()
	if len(names) != 2 {
		t.Errorf("Names() length = %d, want 2", len(names))
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := generator.Config{
		TableName: "users",
		Columns: map[string]generator.ColumnConfig{
			"email": {
				Generator: "email",
				Params:    map[string]any{"domain": "example.com"},
			},
		},
		Count: 100,
	}
	if cfg.TableName != "users" {
		t.Errorf("Config.TableName = %q", cfg.TableName)
	}
	if cfg.Columns["email"].Generator != "email" {
		t.Errorf("ColumnConfig.Generator = %q", cfg.Columns["email"].Generator)
	}
	if cfg.Count != 100 {
		t.Errorf("Config.Count = %d", cfg.Count)
	}
}

func TestColumnConfigDisabled(t *testing.T) {
	cc := generator.ColumnConfig{Generator: "constant", Disabled: true}
	if !cc.Disabled {
		t.Error("ColumnConfig.Disabled should be true")
	}
}
