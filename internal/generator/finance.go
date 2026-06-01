package generator

import (
	"context"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type CurrencyGenerator struct{}

func (g *CurrencyGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return randomPick(currencies), nil
}

type AmountGenerator struct {
	MinCents int64
	MaxCents int64
}

func (g *AmountGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	min := g.MinCents
	max := g.MaxCents
	if min <= 0 {
		min = 100
	}
	if max <= 0 {
		max = 1000000
	}
	cents := randomInt(min, max)
	return float64(cents) / 100, nil
}
