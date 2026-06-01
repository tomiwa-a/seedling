package stream

import (
	"context"
	"sync"
	"testing"

	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

type captureWriter struct {
	mu     sync.Mutex
	tables map[string]writer.Rows
}

func newCaptureWriter() *captureWriter {
	return &captureWriter{tables: make(map[string]writer.Rows)}
}

func (w *captureWriter) WriteTable(ctx context.Context, table *schema.Table, rows writer.Rows) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tables[table.Name] = append(w.tables[table.Name], rows...)
	return nil
}

func (w *captureWriter) Close() error { return nil }

func TestGenerateSingleTable(t *testing.T) {
	sg := New()
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "email", Type: schema.TypeVarchar},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 10},
		},
		TotalCount: 10,
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}

	rows := cw.tables["users"]
	if len(rows) != 10 {
		t.Fatalf("expected 10 rows, got %d", len(rows))
	}

	for i, row := range rows {
		id := row["id"].(int64)
		if id != int64(i+1) {
			t.Errorf("row %d: expected id %d, got %d", i, i+1, id)
		}
		if _, ok := row["email"]; !ok {
			t.Errorf("row %d: missing email", i)
		}
	}
}

func TestGenerateWithFK(t *testing.T) {
	sg := New()
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 5},
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
			}, Count: 20},
		},
		TotalCount: 25,
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}

	if len(cw.tables["users"]) != 5 {
		t.Errorf("expected 5 users, got %d", len(cw.tables["users"]))
	}
	if len(cw.tables["orders"]) != 20 {
		t.Errorf("expected 20 orders, got %d", len(cw.tables["orders"]))
	}

	for _, row := range cw.tables["orders"] {
		userID := row["user_id"].(int64)
		if userID < 1 || userID > 5 {
			t.Errorf("order user_id %d out of range [1,5]", userID)
		}
	}
}

func TestGenerateUniqueColumn(t *testing.T) {
	sg := New()
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "email", Type: schema.TypeVarchar, Unique: true},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 20},
		},
		TotalCount: 20,
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}

	emails := make(map[string]bool)
	for _, row := range cw.tables["users"] {
		email := row["email"].(string)
		if emails[email] {
			t.Errorf("duplicate email: %s", email)
		}
		emails[email] = true
	}
	if len(emails) != 20 {
		t.Errorf("expected 20 unique emails, got %d", len(emails))
	}
}

func TestGenerateEmptyTable(t *testing.T) {
	sg := New()
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name:    "empty",
				Columns: []*schema.Column{},
			}, Count: 0},
		},
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateFKPoolPopulated(t *testing.T) {
	sg := New()

	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "parents",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 3},
		},
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}

	if sg.pool.Count("parents") != 3 {
		t.Errorf("expected 3 PKs in pool, got %d", sg.pool.Count("parents"))
	}
}

func TestGenerateWithHints(t *testing.T) {
	sg := New()
	sg.SetHints("users", map[string]schema.GeneratorHint{
		"email": schema.HintEmail,
	})

	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{
				Name: "users",
				Columns: []*schema.Column{
					{Name: "id", Type: schema.TypeSerial},
					{Name: "email", Type: schema.TypeVarchar},
				},
				Constraints: []*schema.Constraint{
					{Type: schema.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
			}, Count: 5},
		},
	}

	cw := newCaptureWriter()
	err := sg.Generate(context.Background(), p, cw)
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range cw.tables["users"] {
		email := row["email"].(string)
		if len(email) == 0 {
			t.Error("empty email")
		}
	}
}
