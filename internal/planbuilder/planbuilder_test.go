package planbuilder

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestBuildSingleTable(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{Name: "users", SchemaName: "public", Columns: []*schema.Column{
				{Name: "id", Type: schema.TypeSerial},
			}},
		},
	}

	b := New(100)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(p.Tables))
	}
	if p.Tables[0].Count != 100 {
		t.Errorf("expected count 100, got %d", p.Tables[0].Count)
	}
	if p.Tables[0].Table.Name != "users" {
		t.Errorf("expected table 'users', got %q", p.Tables[0].Table.Name)
	}
}

func TestBuildParentChild(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "users", SchemaName: "public",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
				},
			},
			{
				Name: "orders", SchemaName: "public",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "user_id", Type: schema.TypeInteger, FKRef: &schema.FKRef{Table: "users", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "user_id", RefTable: "users", RefColumn: "id"},
				},
			},
		},
	}

	b := New(50)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(p.Tables))
	}

	usersPlan := p.Tables[0]
	if usersPlan.Table.Name != "users" {
		t.Errorf("expected users first, got %q", usersPlan.Table.Name)
	}
	if usersPlan.Count != 50 {
		t.Errorf("expected users count 50, got %d", usersPlan.Count)
	}

	ordersPlan := p.Tables[1]
	if ordersPlan.Table.Name != "orders" {
		t.Errorf("expected orders second, got %q", ordersPlan.Table.Name)
	}
	if ordersPlan.Count < 250 || ordersPlan.Count > 750 {
		t.Errorf("orders count %d outside expected range [250, 750]", ordersPlan.Count)
	}
}

func TestBuildDiamondDependency(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{Name: "a", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
			{Name: "b", Columns: []*schema.Column{
				{Name: "id", Type: schema.TypeSerial},
				{Name: "a_id", FKRef: &schema.FKRef{Table: "a", Column: "id"}},
			}, ForeignKeys: []*schema.ForeignKey{{ColumnName: "a_id", RefTable: "a", RefColumn: "id"}}},
			{Name: "c", Columns: []*schema.Column{
				{Name: "id", Type: schema.TypeSerial},
				{Name: "a_id", FKRef: &schema.FKRef{Table: "a", Column: "id"}},
			}, ForeignKeys: []*schema.ForeignKey{{ColumnName: "a_id", RefTable: "a", RefColumn: "id"}}},
			{Name: "d", Columns: []*schema.Column{
				{Name: "id", Type: schema.TypeSerial},
				{Name: "b_id", FKRef: &schema.FKRef{Table: "b", Column: "id"}},
			}, ForeignKeys: []*schema.ForeignKey{{ColumnName: "b_id", RefTable: "b", RefColumn: "id"}}},
		},
	}

	b := New(10)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Tables) != 4 {
		t.Fatalf("expected 4 tables, got %d", len(p.Tables))
	}

	if p.Tables[0].Table.Name != "a" {
		t.Errorf("expected a first, got %q", p.Tables[0].Table.Name)
	}
}

func TestBuildNoForeignKeys(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{Name: "t1", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
			{Name: "t2", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
			{Name: "t3", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
		},
	}

	b := New(100)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Tables) != 3 {
		t.Fatalf("expected 3 tables, got %d", len(p.Tables))
	}

	for _, tp := range p.Tables {
		if tp.Count != 100 {
			t.Errorf("expected count 100 for %s, got %d", tp.Table.Name, tp.Count)
		}
	}
}

func TestBuildTotalCountNonZero(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{Name: "a", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
		},
	}

	b := New(50)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if p.TotalCount <= 0 {
		t.Errorf("expected positive TotalCount, got %d", p.TotalCount)
	}
}

func TestBuildTableNamesMatch(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{Name: "x", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
			{Name: "y", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
		},
	}

	b := New(10)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	names := p.TableNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestBuildPlanSeed(t *testing.T) {
	s := &schema.Schema{
		Tables: []*schema.Table{
			{Name: "t", Columns: []*schema.Column{{Name: "id", Type: schema.TypeSerial}}},
		},
	}

	b := New(100)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestBuildEmptySchema(t *testing.T) {
	s := &schema.Schema{Name: "empty"}
	b := New(100)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tables) != 0 {
		t.Errorf("expected 0 tables, got %d", len(p.Tables))
	}
}
