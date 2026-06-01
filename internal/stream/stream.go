package stream

import (
	"context"
	"fmt"

	genlib "github.com/tomiwa-a/seedling/internal/generator"
	"github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

type Generator struct {
	pool    *genlib.FKPool
	tracker *genlib.UniqueTracker
	hints   map[string]map[string]schema.GeneratorHint
}

func New() *Generator {
	return &Generator{
		pool:    genlib.NewFKPool(),
		tracker: genlib.NewUniqueTracker(),
		hints:   make(map[string]map[string]schema.GeneratorHint),
	}
}

func (g *Generator) SetHints(table string, hints map[string]schema.GeneratorHint) {
	g.hints[table] = hints
}

func (g *Generator) Generate(ctx context.Context, p *plan.Plan, w writer.Writer) error {
	for _, tp := range p.Tables {
		if err := g.generateTable(ctx, tp, w); err != nil {
			return fmt.Errorf("generate %s: %w", tp.Table.Name, err)
		}
	}
	return w.Close()
}

func (g *Generator) generateTable(ctx context.Context, tp *plan.TablePlan, w writer.Writer) error {
	gens, err := g.resolveGenerators(tp.Table)
	if err != nil {
		return err
	}

	pkColumns := findPKColumns(tp.Table)
	batchSize := 1000
	var batch writer.Rows

	for i := int64(0); i < int64(tp.Count); i++ {
		row := make(writer.Row)
		rc := &rowCtx{
			index:  i,
			table:  tp.Table.Name,
			values: row,
		}

		for _, col := range tp.Table.Columns {
			if col.Unique && col.FKRef != nil {
				consumerKey := tp.Table.Name + "." + col.Name
				val, err := g.pool.Consume(col.FKRef.Table, consumerKey)
				if err != nil {
					return fmt.Errorf("consume FK for %s: %w", col.Name, err)
				}
				row[col.Name] = val
				continue
			}

			gen, ok := gens[col.Name]
			if !ok {
				return fmt.Errorf("no generator for column %s", col.Name)
			}

			val, err := gen.Generate(ctx, rc)
			if err != nil {
				return fmt.Errorf("generate %s: %w", col.Name, err)
			}

			if col.Unique {
				key := tp.Table.Name + "." + col.Name
				resolved, err := g.tracker.Generate(key, 100, func() any {
					g, _ := gens[col.Name]
					v, _ := g.Generate(ctx, rc)
					return v
				})
				if err != nil {
					return fmt.Errorf("unique constraint %s: %w", key, err)
				}
				val = resolved
			}

			row[col.Name] = val
		}

		for _, pk := range pkColumns {
			if val, ok := row[pk]; ok {
				g.pool.Add(tp.Table.Name, val)
			}
		}

		batch = append(batch, row)

		if len(batch) >= batchSize {
			if err := w.WriteTable(ctx, tp.Table, batch); err != nil {
				return err
			}
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err := w.WriteTable(ctx, tp.Table, batch); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) resolveGenerators(tbl *schema.Table) (map[string]generator.Generator, error) {
	tableHints := g.hints[tbl.Name]
	if tableHints == nil {
		tableHints = make(map[string]schema.GeneratorHint)
	}
	return genlib.ResolveGenerators(tbl.Columns, tableHints, g.pool)
}

func findPKColumns(tbl *schema.Table) []string {
	for _, c := range tbl.Constraints {
		if c.Type == schema.ConstraintPrimaryKey {
			return c.Columns
		}
	}
	for _, col := range tbl.Columns {
		if col.Type == schema.TypeSerial || col.Type == schema.TypeBigSerial {
			return []string{col.Name}
		}
	}
	return nil
}

type rowCtx struct {
	index  int64
	table  string
	values map[string]any
}

func (r *rowCtx) Column(name string) (any, bool) {
	v, ok := r.values[name]
	return v, ok
}

func (r *rowCtx) RowIndex() int64 {
	return r.index
}

func (r *rowCtx) TableName() string {
	return r.table
}
