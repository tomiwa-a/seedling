package planbuilder

import (
	"context"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestBuildCircularFK(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "employees",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "name", Type: schema.TypeVarchar},
					{Name: "manager_id", Type: schema.TypeInteger, Nullable: true,
						FKRef: &schema.FKRef{Table: "employees", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "manager_id", RefTable: "employees", RefColumn: "id"},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			},
		},
	}

	b := New(10)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if p.CircularGroup == nil {
		t.Fatal("expected CircularGroup to be set")
	}

	if len(p.CircularGroup.CycleEdges) == 0 {
		t.Error("expected cycle edges")
	}

	if len(p.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(p.Tables))
	}
}

func TestBuildMutualCircularFK(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "table_a",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "b_id", Type: schema.TypeInteger, Nullable: true,
						FKRef: &schema.FKRef{Table: "table_b", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "b_id", RefTable: "table_b", RefColumn: "id"},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			},
			{
				Name: "table_b",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "a_id", Type: schema.TypeInteger, Nullable: true,
						FKRef: &schema.FKRef{Table: "table_a", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "a_id", RefTable: "table_a", RefColumn: "id"},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			},
		},
	}

	b := New(10)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	if p.CircularGroup == nil {
		t.Fatal("expected CircularGroup for mutual circular FK")
	}

	if len(p.CircularGroup.Pass1Tables) == 0 {
		t.Error("expected non-empty Pass1Tables")
	}
	if len(p.CircularGroup.Pass2Tables) == 0 {
		t.Error("expected non-empty Pass2Tables")
	}
}

func TestBuildCircularWithNonCircular(t *testing.T) {
	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			},
			{
				Name: "employees",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "manager_id", Type: schema.TypeInteger, Nullable: true,
						FKRef: &schema.FKRef{Table: "employees", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "manager_id", RefTable: "employees", RefColumn: "id"},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			},
		},
	}

	b := New(10)
	p, err := b.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}

	usersFound := false
	employeesFound := false
	for _, tp := range p.Tables {
		if tp.Table.Name == "users" {
			usersFound = true
		}
		if tp.Table.Name == "employees" {
			employeesFound = true
		}
	}

	if !usersFound {
		t.Error("users table missing from plan")
	}
	if !employeesFound {
		t.Error("employees table missing from plan")
	}
}
