package generator

import (
	"context"
	"math/rand"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type CurrencyGenerator struct {
	rnd *rand.Rand
}

func (g *CurrencyGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *CurrencyGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	return randomPick(rnd, currencies), nil
}

type AmountGenerator struct {
	MinCents int64
	MaxCents int64
	rnd      *rand.Rand
}

func (g *AmountGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *AmountGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	min := g.MinCents
	max := g.MaxCents
	if min <= 0 {
		min = 100
	}
	if max <= 0 {
		max = 1000000
	}
	cents := randomInt(rnd, min, max)
	return float64(cents) / 100, nil
}
