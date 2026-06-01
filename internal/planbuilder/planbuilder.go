package planbuilder

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/tomiwa-a/seedling/pkg/generator"
	"github.com/tomiwa-a/seedling/pkg/plan"
	"github.com/tomiwa-a/seedling/pkg/schema"
)

type Builder struct {
	rootCount int
	seed      int64
}

func New(rootCount int) *Builder {
	return &Builder{
		rootCount: rootCount,
	}
}

func (b *Builder) Build(ctx context.Context, s *schema.Schema, configs []generator.Config) (*plan.Plan, error) {
	tableMap := make(map[string]*schema.Table)
	for i := range s.Tables {
		tableMap[s.Tables[i].Name] = s.Tables[i]
	}

	deps := buildDependencyGraph(s.Tables)

	ordered, err := topologicalSort(s.Tables, deps)
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}

	uniqueFKs := buildUniqueFKMap(s.Tables)
	counts := computeCounts(ordered, deps, uniqueFKs, b.rootCount)

	var tablePlans []*plan.TablePlan
	var totalCount int64

	for _, tbl := range ordered {
		count := counts[tbl.Name]
		totalCount += int64(count)

		tablePlans = append(tablePlans, &plan.TablePlan{
			Table: tbl,
			Count: count,
		})
	}

	return &plan.Plan{
		Tables:     tablePlans,
		TotalCount: totalCount,
		Seed:       b.seed,
		BatchSize:  1000,
		CreatedAt:  time.Now(),
	}, nil
}

func buildDependencyGraph(tables []*schema.Table) map[string][]string {
	tableSet := make(map[string]bool)
	for _, t := range tables {
		tableSet[t.Name] = true
	}

	deps := make(map[string][]string)
	for _, t := range tables {
		for _, fk := range t.ForeignKeys {
			if tableSet[fk.RefTable] {
				deps[t.Name] = append(deps[t.Name], fk.RefTable)
			}
		}
	}
	return deps
}

func topologicalSort(tables []*schema.Table, deps map[string][]string) ([]*schema.Table, error) {
	tableMap := make(map[string]*schema.Table)
	for i := range tables {
		tableMap[tables[i].Name] = tables[i]
	}

	inDegree := make(map[string]int)
	for _, t := range tables {
		inDegree[t.Name] = len(deps[t.Name])
	}

	var queue []string
	for _, t := range tables {
		if inDegree[t.Name] == 0 {
			queue = append(queue, t.Name)
		}
	}

	var sorted []*schema.Table
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		sorted = append(sorted, tableMap[name])

		for child, parents := range deps {
			for _, parent := range parents {
				if parent == name {
					inDegree[child]--
					if inDegree[child] == 0 {
						queue = append(queue, child)
					}
				}
			}
		}
	}

	if len(sorted) != len(tables) {
		return nil, fmt.Errorf("circular dependency detected in schema")
	}

	return sorted, nil
}

func buildUniqueFKMap(tables []*schema.Table) map[string]struct{} {
	uniqueFK := make(map[string]struct{})
	for _, t := range tables {
		fkCols := make(map[string]bool)
		for _, fk := range t.ForeignKeys {
			fkCols[fk.ColumnName] = true
		}
		for _, c := range t.Columns {
			if c.Unique && fkCols[c.Name] {
				uniqueFK[t.Name] = struct{}{}
				break
			}
		}
	}
	return uniqueFK
}

func computeCounts(ordered []*schema.Table, deps map[string][]string, uniqueFKs map[string]struct{}, rootCount int) map[string]int {
	counts := make(map[string]int)
	tableSet := make(map[string]bool)
	for _, t := range ordered {
		tableSet[t.Name] = true
	}

	for _, t := range ordered {
		parents := deps[t.Name]
		if len(parents) == 0 {
			counts[t.Name] = rootCount
		} else {
			maxParentCount := 0
			for _, p := range parents {
				if tableSet[p] {
					if counts[p] > maxParentCount {
						maxParentCount = counts[p]
					}
				}
			}
			if maxParentCount == 0 {
				maxParentCount = rootCount
			}

			if _, hasUniqueFK := uniqueFKs[t.Name]; hasUniqueFK {
				counts[t.Name] = maxParentCount
			} else {
				multiplier := 5 + int(cryptoRandIntn(11))
				counts[t.Name] = maxParentCount * multiplier
			}
		}
	}

	return counts
}

func cryptoRandIntn(n int) int64 {
	if n <= 0 {
		return 0
	}
	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return val.Int64()
}
