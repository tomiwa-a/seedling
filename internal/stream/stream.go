package stream

import (
	"context"
	"fmt"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"

	genlib "github.com/tomiwa-a/seedling/internal/generator"
	"github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
	"github.com/tomiwa-a/seedling/pkg/writer"
)

type Generator struct {
	pool       *genlib.FKPool
	tracker    *genlib.UniqueTracker
	hints      map[string]map[string]schema.GeneratorHint
	seed       uint64
	onProgress ProgressFunc
	parallel   bool
}

func New() *Generator {
	return &Generator{
		pool:    genlib.NewFKPool(),
		tracker: genlib.NewUniqueTracker(),
		hints:   make(map[string]map[string]schema.GeneratorHint),
		seed:    0,
		parallel: false,
	}
}

func (g *Generator) SetProgress(fn ProgressFunc) {
	g.onProgress = fn
}

func (g *Generator) SetParallel(p bool) {
	g.parallel = p
}

func (g *Generator) SetSeed(seed uint64) {
	g.seed = seed
	if seed != 0 {
		deriver := genlib.NewSeedDeriver(seed)
		rnd := genlib.NewSeededRand(deriver.TableSeed("_pool"))
		g.pool.SetRand(rnd)
	}
}

func (g *Generator) SetHints(table string, hints map[string]schema.GeneratorHint) {
	g.hints[table] = hints
}

func (g *Generator) SetHint(table, column string, hint schema.GeneratorHint) {
	if g.hints[table] == nil {
		g.hints[table] = make(map[string]schema.GeneratorHint)
	}
	g.hints[table][column] = hint
}

func (g *Generator) Generate(ctx context.Context, p *plan.Plan, w writer.Writer) error {
	if p.CircularGroup != nil {
		return g.generateCircular(ctx, p, w)
	}

	if g.parallel && len(p.Tables) > 1 {
		return g.generateParallel(ctx, p, w)
	}

	for _, tp := range p.Tables {
		if err := g.generateTable(ctx, tp, w); err != nil {
			return fmt.Errorf("generate %s: %w", tp.Table.Name, err)
		}
	}
	return w.Close()
}

func (g *Generator) generateParallel(ctx context.Context, p *plan.Plan, w writer.Writer) error {
	levels := computeLevels(p)

	for _, level := range levels {
		if len(level) == 1 {
			if err := g.generateTable(ctx, level[0], w); err != nil {
				return fmt.Errorf("generate %s: %w", level[0].Table.Name, err)
			}
			continue
		}

		eg, ctx := errgroup.WithContext(ctx)
		for _, tp := range level {
			tp := tp
			eg.Go(func() error {
				return g.generateTable(ctx, tp, w)
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return w.Close()
}

func computeLevels(p *plan.Plan) [][]*plan.TablePlan {
	tableSet := make(map[string]bool)
	for _, tp := range p.Tables {
		tableSet[tp.Table.Name] = true
	}

	deps := make(map[string][]string)
	for _, tp := range p.Tables {
		for _, fk := range tp.Table.ForeignKeys {
			if tableSet[fk.RefTable] {
				deps[tp.Table.Name] = append(deps[tp.Table.Name], fk.RefTable)
			}
		}
	}

	inDegree := make(map[string]int)
	for _, tp := range p.Tables {
		inDegree[tp.Table.Name] = len(deps[tp.Table.Name])
	}

	tableMap := make(map[string]*plan.TablePlan)
	for _, tp := range p.Tables {
		tableMap[tp.Table.Name] = tp
	}

	var levels [][]*plan.TablePlan
	remaining := len(p.Tables)

	for remaining > 0 {
		var level []string
		for name, deg := range inDegree {
			if deg == 0 {
				level = append(level, name)
			}
		}

		if len(level) == 0 {
			break
		}

		sort.Strings(level)

		var tablePlanLevel []*plan.TablePlan
		for _, name := range level {
			tablePlanLevel = append(tablePlanLevel, tableMap[name])
		}
		levels = append(levels, tablePlanLevel)

		for _, name := range level {
			remaining--
			inDegree[name] = -1
			for child, parents := range deps {
				for _, parent := range parents {
					if parent == name {
						inDegree[child]--
					}
				}
			}
		}
	}

	return levels
}

func (g *Generator) generateCircular(ctx context.Context, p *plan.Plan, w writer.Writer) error {
	cg := p.CircularGroup
	pass1Set := make(map[string]bool)
	for _, name := range cg.Pass1Tables {
		pass1Set[name] = true
	}

	pass2Set := make(map[string]bool)
	for _, name := range cg.Pass2Tables {
		pass2Set[name] = true
	}

	selfRefTables := make(map[string]bool)
	for _, tp := range p.Tables {
		for _, fk := range tp.Table.ForeignKeys {
			if fk.RefTable == tp.Table.Name {
				selfRefTables[tp.Table.Name] = true
				break
			}
		}
	}

	for _, tp := range p.Tables {
		if pass1Set[tp.Table.Name] {
			if selfRefTables[tp.Table.Name] {
				if err := g.generateCircularTable(ctx, tp, w); err != nil {
					return fmt.Errorf("generate %s: %w", tp.Table.Name, err)
				}
			} else {
				if err := g.generateTable(ctx, tp, w); err != nil {
					return fmt.Errorf("generate %s: %w", tp.Table.Name, err)
				}
			}
		}
	}

	for _, tp := range p.Tables {
		if pass2Set[tp.Table.Name] {
			if err := g.generateCircularTable(ctx, tp, w); err != nil {
				return fmt.Errorf("generate %s: %w", tp.Table.Name, err)
			}
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
	start := time.Now()
	var written int64

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
			written += int64(len(batch))
			batch = nil

			if g.onProgress != nil {
				g.onProgress(calcProgress(tp.Table.Name, written, int64(tp.Count), time.Since(start)))
			}
		}
	}

	if len(batch) > 0 {
		if err := w.WriteTable(ctx, tp.Table, batch); err != nil {
			return err
		}
		written += int64(len(batch))
	}

	if g.onProgress != nil {
		g.onProgress(calcProgress(tp.Table.Name, written, int64(tp.Count), time.Since(start)))
	}

	return nil
}

func (g *Generator) generateCircularTable(ctx context.Context, tp *plan.TablePlan, w writer.Writer) error {
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
			if col.FKRef != nil && col.Nullable && col.FKRef.Table == tp.Table.Name {
				if g.pool.Count(col.FKRef.Table) > 0 {
					val, err := g.pool.Pick(col.FKRef.Table)
					if err == nil {
						row[col.Name] = val
						continue
					}
				}
				row[col.Name] = nil
				continue
			}

			if col.Unique && col.FKRef != nil {
				consumerKey := tp.Table.Name + "." + col.Name
				val, err := g.pool.Consume(col.FKRef.Table, consumerKey)
				if err != nil {
					if col.Nullable {
						row[col.Name] = nil
						continue
					}
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
	return genlib.ResolveGenerators(tbl.Columns, tableHints, g.pool, g.seed, tbl.Name)
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
