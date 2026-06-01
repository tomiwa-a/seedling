package generator

import (
	"context"
	"fmt"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type SelfFKGenerator struct {
	TableName    string
	ColumnName   string
	pool         *FKPool
	depth        int
	maxDepth     int
	generatedIDs []any
	rnd          *rand.Rand
}

func NewSelfFKGenerator(tableName, columnName string, pool *FKPool, maxDepth int) *SelfFKGenerator {
	return &SelfFKGenerator{
		TableName:  tableName,
		ColumnName: columnName,
		pool:       pool,
		maxDepth:   maxDepth,
		rnd:        rand.New(rand.NewSource(0)),
	}
}

func (g *SelfFKGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *SelfFKGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	if g.depth >= g.maxDepth {
		return nil, nil
	}

	g.depth++

	if g.pool.Count(g.TableName) == 0 {
		return nil, nil
	}

	val, err := g.pool.Pick(g.TableName)
	if err != nil {
		return nil, nil
	}

	g.generatedIDs = append(g.generatedIDs, val)
	return val, nil
}

func (g *SelfFKGenerator) ResetDepth() {
	g.depth = 0
}

func (g *SelfFKGenerator) String() string {
	return fmt.Sprintf("SelfFK(%s.%s, maxDepth=%d)", g.TableName, g.ColumnName, g.maxDepth)
}
