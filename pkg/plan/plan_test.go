package plan_test

import (
	"context"
	"testing"
	"time"

	"github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestPlanBuilderFuncAdapter(t *testing.T) {
	var called bool
	pb := plan.PlanBuilderFunc(func(ctx context.Context, s *schema.Schema, configs []generator.Config) (*plan.Plan, error) {
		called = true
		if len(s.Tables) != 1 {
			t.Errorf("tables length = %d", len(s.Tables))
		}
		return &plan.Plan{
			Tables: []*plan.TablePlan{
				{Table: s.Tables[0], Count: 100},
			},
			TotalCount: 100,
			Seed:       42,
		}, nil
	})

	s := &schema.Schema{
		Tables: []*schema.Table{
			{Name: "users"},
		},
	}

	p, err := pb.Build(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("PlanBuilderFunc was not called")
	}
	if p.TotalCount != 100 {
		t.Errorf("TotalCount = %d, want 100", p.TotalCount)
	}
	if p.Seed != 42 {
		t.Errorf("Seed = %d, want 42", p.Seed)
	}
}

func TestPlanTableNames(t *testing.T) {
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{Name: "users"}, Count: 100},
			{Table: &schema.Table{Name: "orders"}, Count: 500},
		},
		TotalCount: 600,
	}

	names := p.TableNames()
	expected := []string{"users", "orders"}
	if len(names) != len(expected) {
		t.Fatalf("TableNames() = %v, want %v", names, expected)
	}
	for i := range expected {
		if names[i] != expected[i] {
			t.Errorf("TableNames()[%d] = %q, want %q", i, names[i], expected[i])
		}
	}
}

func TestPlanEstimateRowCount(t *testing.T) {
	p := &plan.Plan{
		Tables: []*plan.TablePlan{
			{Table: &schema.Table{Name: "users"}, Count: 100},
			{Table: &schema.Table{Name: "orders"}, Count: 500},
		},
		TotalCount: 600,
	}
	if got := p.EstimateRowCount(); got != 600 {
		t.Errorf("EstimateRowCount() = %d, want 600", got)
	}
}

func TestPlanEmpty(t *testing.T) {
	p := &plan.Plan{}
	if got := p.EstimateRowCount(); got != 0 {
		t.Errorf("EstimateRowCount() = %d, want 0", got)
	}
	if len(p.TableNames()) != 0 {
		t.Error("expected empty TableNames")
	}
}

func TestTablePlanStruct(t *testing.T) {
	tp := &plan.TablePlan{
		Table: &schema.Table{Name: "users"},
		Count: 1000,
		Pass:  1,
	}
	if tp.Table.Name != "users" {
		t.Errorf("TablePlan.Table.Name = %q", tp.Table.Name)
	}
	if tp.Count != 1000 {
		t.Errorf("TablePlan.Count = %d", tp.Count)
	}
	if tp.Pass != 1 {
		t.Errorf("TablePlan.Pass = %d", tp.Pass)
	}
}

func TestPlanCreatedAtSet(t *testing.T) {
	p := plan.Plan{
		CreatedAt: time.Now(),
	}
	if p.CreatedAt.IsZero() {
		t.Error("Plan.CreatedAt should not be zero")
	}
}
