package generator

import (
	"context"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type RandomIntGenerator struct {
	Min int64
	Max int64
	rnd *rand.Rand
}

func (g *RandomIntGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *RandomIntGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomInt(rnd, g.Min, g.Max), nil
}

type NumericGenerator struct {
	Min int64
	Max int64
	rnd *rand.Rand
}

func (g *NumericGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *NumericGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	cents := randomInt(rnd, g.Min, g.Max)
	return float64(cents) / 100, nil
}

type FloatRangeGenerator struct {
	Min float64
	Max float64
	rnd *rand.Rand
}

func (g *FloatRangeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *FloatRangeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	scale := g.Max - g.Min
	return g.Min + (rnd.Float64() * scale), nil
}

type ConstantGenerator struct {
	Value any
}

func (g *ConstantGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return g.Value, nil
}

type WeightedChoiceGenerator struct {
	Weights []WeightedChoice
	rnd     *rand.Rand
}

type WeightedChoice struct {
	Label  string
	Weight int
}

func (g *WeightedChoiceGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *WeightedChoiceGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	total := 0
	for _, w := range g.Weights {
		total += w.Weight
	}
	r := rnd.Intn(total)
	for _, w := range g.Weights {
		r -= w.Weight
		if r < 0 {
			return w.Label, nil
		}
	}
	return g.Weights[len(g.Weights)-1].Label, nil
}
