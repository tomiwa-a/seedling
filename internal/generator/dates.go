package generator

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	gen "github.com/tomiwa-a/seedling/pkg/generator"
)

type DateGenerator struct {
	MinYear int
	MaxYear int
	rnd     *rand.Rand
}

func (g *DateGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *DateGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	yearDelta := g.MaxYear - g.MinYear
	y := rnd.Intn(yearDelta + 1)
	m := rnd.Intn(12)
	d := rnd.Intn(28)
	return fmt.Sprintf("%04d-%02d-%02d",
		g.MinYear+y,
		m+1,
		d+1,
	), nil
}

type TimestampGenerator struct {
	rnd *rand.Rand
}

func (g *TimestampGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *TimestampGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	y := rnd.Intn(3)
	m := rnd.Intn(12)
	d := rnd.Intn(28)
	h := rnd.Intn(24)
	mi := rnd.Intn(60)
	s := rnd.Intn(60)
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		2024+y, m+1, d+1,
		h, mi, s,
	), nil
}

type NowGenerator struct{}

func (g *NowGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z"), nil
}

type TimeAgoGenerator struct {
	MaxDays int
	rnd     *rand.Rand
}

func (g *TimeAgoGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *TimeAgoGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	maxDays := g.MaxDays
	if maxDays <= 0 {
		maxDays = 365
	}
	daysAgo := rnd.Intn(maxDays)
	t := time.Now().UTC().AddDate(0, 0, -daysAgo)
	return t.Format("2006-01-02T15:04:05Z"), nil
}

type BusinessDaysGenerator struct {
	MinYear int
	MaxYear int
	rnd     *rand.Rand
}

func (g *BusinessDaysGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *BusinessDaysGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	yearDelta := g.MaxYear - g.MinYear
	y := rnd.Intn(yearDelta + 1)
	m := rnd.Intn(12)
	d := rnd.Intn(28)
	t := time.Date(g.MinYear+y, time.Month(m+1), d+1, 0, 0, 0, 0, time.UTC)
	for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	return t.Format("2006-01-02"), nil
}

type TimeGenerator struct {
	rnd *rand.Rand
}

func (g *TimeGenerator) SetRand(rnd *rand.Rand) { g.rnd = rnd }

func (g *TimeGenerator) Generate(ctx context.Context, row gen.RowContext) (any, error) {
	rnd := g.rnd
	if rnd == nil {
		rnd = rand.New(rand.NewSource(0))
	}
	h := rnd.Intn(24)
	m := rnd.Intn(60)
	s := rnd.Intn(60)
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s), nil
}
