package generator

import (
	"context"
	"crypto/rand"
	"math/big"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type RandomIntGenerator struct {
	Min int64
	Max int64
}

func (g *RandomIntGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomInt(g.Min, g.Max), nil
}

type NumericGenerator struct {
	Min int64
	Max int64
}

func (g *NumericGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	cents := randomInt(g.Min, g.Max)
	return float64(cents) / 100, nil
}

type FloatRangeGenerator struct {
	Min float64
	Max float64
}

func (g *FloatRangeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, err
	}
	scale := g.Max - g.Min
	return g.Min + (float64(n.Int64()) / 1000000.0 * scale), nil
}

type ConstantGenerator struct {
	Value any
}

func (g *ConstantGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return g.Value, nil
}

type WeightedChoiceGenerator struct {
	Weights []WeightedChoice
}

type WeightedChoice struct {
	Label  string
	Weight int
}

func (g *WeightedChoiceGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	total := 0
	for _, w := range g.Weights {
		total += w.Weight
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(total)))
	r := int(n.Int64())
	for _, w := range g.Weights {
		r -= w.Weight
		if r < 0 {
			return w.Label, nil
		}
	}
	return g.Weights[len(g.Weights)-1].Label, nil
}
