package plan

import (
	"context"
	"time"

	"github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type TablePlan struct {
	Table      *schema.Table                     `json:"table" yaml:"table"`
	Count      int                               `json:"count" yaml:"count"`
	Generators map[string]generator.Generator    `json:"-" yaml:"-"`
	Pass       int                               `json:"pass,omitempty" yaml:"pass,omitempty"`
}

type Plan struct {
	Tables     []*TablePlan `json:"tables" yaml:"tables"`
	TotalCount int64        `json:"total_count" yaml:"total_count"`
	Seed       int64        `json:"seed" yaml:"seed"`
	BatchSize  int          `json:"batch_size" yaml:"batch_size"`
	CreatedAt  time.Time    `json:"created_at" yaml:"created_at"`
}

type PlanBuilder interface {
	Build(ctx context.Context, s *schema.Schema, configs []generator.Config) (*Plan, error)
}

type PlanBuilderFunc func(ctx context.Context, s *schema.Schema, configs []generator.Config) (*Plan, error)

func (f PlanBuilderFunc) Build(ctx context.Context, s *schema.Schema, configs []generator.Config) (*Plan, error) {
	return f(ctx, s, configs)
}

func (p *Plan) EstimateRowCount() int64 {
	return p.TotalCount
}

func (p *Plan) TableNames() []string {
	names := make([]string, len(p.Tables))
	for i, tp := range p.Tables {
		names[i] = tp.Table.Name
	}
	return names
}
