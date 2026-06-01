package stream

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

func TestDeterminismSameSeed(t *testing.T) {
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "email", Type: schema.TypeVarchar, Hint: schema.HintEmail},
					{Name: "name", Type: schema.TypeVarchar, Hint: schema.HintName},
					{Name: "active", Type: schema.TypeBoolean},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 20},
			{Table: &schema.Table{
				Name: "orders",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "user_id", Type: schema.TypeInteger,
						FKRef: &schema.FKRef{Table: "users", Column: "id"}},
				},
				ForeignKeys: []*schema.ForeignKey{
					{ColumnName: "user_id", RefTable: "users", RefColumn: "id"},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 30},
		},
		TotalCount: 50,
	}

	var buf1, buf2 bytes.Buffer
	w1 := writer.NewWriterFunc(func(ctx context.Context, table *schema.Table, rows writer.Rows) error {
		for _, row := range rows {
			for _, col := range table.ColumnNames() {
				buf1.WriteString(fmt.Sprintf("%v", row[col]))
			}
		}
		return nil
	})
	w2 := writer.NewWriterFunc(func(ctx context.Context, table *schema.Table, rows writer.Rows) error {
		for _, row := range rows {
			for _, col := range table.ColumnNames() {
				buf2.WriteString(fmt.Sprintf("%v", row[col]))
			}
		}
		return nil
	})

	sg1 := New()
	sg1.SetSeed(42)
	if err := sg1.Generate(context.Background(), p, w1); err != nil {
		t.Fatal(err)
	}

	sg2 := New()
	sg2.SetSeed(42)
	if err := sg2.Generate(context.Background(), p, w2); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("same seed should produce identical output")
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "t",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "val", Type: schema.TypeVarchar},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 10},
		},
	}

	var buf1, buf2 bytes.Buffer
	w1 := writer.NewWriterFunc(func(ctx context.Context, table *schema.Table, rows writer.Rows) error {
		for _, row := range rows {
			buf1.WriteString(fmt.Sprintf("%v", row["val"]))
		}
		return nil
	})
	w2 := writer.NewWriterFunc(func(ctx context.Context, table *schema.Table, rows writer.Rows) error {
		for _, row := range rows {
			buf2.WriteString(fmt.Sprintf("%v", row["val"]))
		}
		return nil
	})

	sg1 := New()
	sg1.SetSeed(42)
	sg1.Generate(context.Background(), p, w1)

	sg2 := New()
	sg2.SetSeed(99)
	sg2.Generate(context.Background(), p, w2)

	if bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("different seeds should produce different output")
	}
}
